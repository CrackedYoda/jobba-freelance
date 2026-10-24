import "server-only";

import { randomUUID } from "node:crypto";
import type { CreateProposalInput, ListProposalsInput } from "@/features/proposals/schemas";
import { decodeCursor, encodeCursor } from "@/server/db/cursor";
import { getDatabase } from "@/server/db/client";
import { AppError } from "@/server/db/errors";
import { findContractByProposal, insertFundingPendingContract } from "@/server/repositories/contracts";
import { getProjectById } from "@/server/repositories/projects";
import { getProposalById, insertProposal, listProjectProposals, withdrawPendingProposal } from "@/server/repositories/proposals";
import type { Actor } from "@/server/repositories/profiles";
import { serializeContract, serializeProposal } from "@/server/services/serialize";

export async function createProposal(actor: Actor, projectId: string, input: CreateProposalInput) {
  if (!actor.roles.includes("freelancer")) throw new AppError("FORBIDDEN", "The freelancer role is required.", 403);
  const db = getDatabase();
  const project = await getProjectById(db, projectId);
  if (!project) throw new AppError("PROJECT_NOT_FOUND", "Project not found.", 404);
  if (project.status !== "open") throw new AppError("PROJECT_NOT_OPEN", "This project is not accepting proposals.", 409);
  if (project.client_id === actor.profileId) throw new AppError("OWN_PROJECT", "You cannot propose on your own project.", 403);
  return serializeProposal(await insertProposal(db, projectId, actor.profileId, input) as Record<string, unknown>);
}

export async function readProjectProposals(actor: Actor, projectId: string, input: ListProposalsInput) {
  const db = getDatabase();
  const project = await getProjectById(db, projectId);
  if (!project) throw new AppError("PROJECT_NOT_FOUND", "Project not found.", 404);
  if (project.client_id !== actor.profileId && !actor.roles.includes("admin")) {
    throw new AppError("FORBIDDEN", "Only the project client can review its proposals.", 403);
  }
  const rows = await listProjectProposals(db, projectId, input, decodeCursor(input.cursor));
  const hasMore = rows.length > input.limit;
  const page = rows.slice(0, input.limit);
  const last = page.at(-1) as Record<string, unknown> | undefined;
  return {
    items: page.map((row) => serializeProposal(row as Record<string, unknown>)),
    nextCursor: hasMore && last ? encodeCursor({ createdAt: String(last.created_at), id: String(last.id) }) : null,
  };
}

export async function withdrawProposal(actor: Actor, proposalId: string) {
  if (!actor.roles.includes("freelancer")) throw new AppError("FORBIDDEN", "The freelancer role is required.", 403);
  const db = getDatabase();
  const proposal = await getProposalById(db, proposalId);
  if (!proposal) throw new AppError("PROPOSAL_NOT_FOUND", "Proposal not found.", 404);
  if (proposal.freelancer_id !== actor.profileId) throw new AppError("FORBIDDEN", "You cannot withdraw this proposal.", 403);
  if (proposal.status !== "pending") throw new AppError("INVALID_PROPOSAL_STATE", "Only pending proposals can be withdrawn.", 409);
  const updated = await withdrawPendingProposal(db, proposalId, actor.profileId, Number(proposal.version));
  if (!updated) throw new AppError("PROPOSAL_CONFLICT", "The proposal changed before it could be withdrawn.", 409);
  return serializeProposal(updated as Record<string, unknown>);
}

export async function prepareProposalFunding(actor: Actor, proposalId: string) {
  if (!actor.roles.includes("client")) throw new AppError("FORBIDDEN", "The client role is required.", 403);
  const db = getDatabase();
  const proposal = await getProposalById(db, proposalId);
  if (!proposal) throw new AppError("PROPOSAL_NOT_FOUND", "Proposal not found.", 404);
  const project = await getProjectById(db, proposal.project_id as string);
  if (!project) throw new AppError("PROJECT_NOT_FOUND", "Project not found.", 404);
  if (project.client_id !== actor.profileId) throw new AppError("FORBIDDEN", "Only the project client can select a proposal.", 403);
  if (project.status !== "open") {
    const existing = await findContractByProposal(db, proposalId);
    if (existing) return serializeContract(existing as Record<string, unknown>);
    throw new AppError("PROJECT_NOT_OPEN", "This project is no longer open.", 409);
  }
  if (proposal.status !== "pending") throw new AppError("INVALID_PROPOSAL_STATE", "Only pending proposals can be selected.", 409);
  const existing = await findContractByProposal(db, proposalId);
  if (existing) return serializeContract(existing as Record<string, unknown>);
  try {
    const contract = await insertFundingPendingContract(db, {
      id: randomUUID(),
      projectId: project.id as string,
      proposalId: proposal.id as string,
      clientId: actor.profileId,
      freelancerId: proposal.freelancer_id as string,
      amountAtomic: String(proposal.amount_atomic),
      currencyMint: project.currency_mint as string,
    });
    return serializeContract(contract as Record<string, unknown>);
  } catch (error) {
    if (error instanceof AppError && error.code === "CONFLICT") {
      const raced = await findContractByProposal(db, proposalId);
      if (raced) return serializeContract(raced as Record<string, unknown>);
    }
    throw error;
  }
}

export async function rejectProposal(actor: Actor, proposalId: string) {
  const db = getDatabase();
  const proposal = await getProposalById(db, proposalId);
  if (!proposal) throw new AppError("PROPOSAL_NOT_FOUND", "Proposal not found.", 404);
  const project = await getProjectById(db, proposal.project_id as string);
  if (!project || project.client_id !== actor.profileId) throw new AppError("FORBIDDEN", "Only the project client can reject a proposal.", 403);
  if (proposal.status !== "pending") throw new AppError("INVALID_PROPOSAL_STATE", "Only pending proposals can be rejected.", 409);
  const contract = await findContractByProposal(db, proposalId);
  if (contract) throw new AppError("PROPOSAL_SELECTED", "This proposal already has a contract. Abandon it first.", 409);
  const { rowCount } = await db.query("update proposals set status = 'rejected', version = version + 1 where id = $1 and status = 'pending'", [proposalId]);
  if (!rowCount) throw new AppError("PROPOSAL_CONFLICT", "The proposal changed before it could be rejected.", 409);
  return serializeProposal(await getProposalById(db, proposalId) as Record<string, unknown>);
}

/** Removes a contract that was selected but never funded, so the client can pick another proposal. */
export async function abandonUnfundedContract(actor: Actor, contractId: string) {
  const db = getDatabase();
  const conn = await db.connect();
  try {
    await conn.query("begin");
    const { rows } = await conn.query<{ client_id: string; status: string }>("select client_id, status from contracts where id = $1 for update", [contractId]);
    const contract = rows[0];
    if (!contract) throw new AppError("CONTRACT_NOT_FOUND", "Contract not found.", 404);
    if (contract.client_id !== actor.profileId) throw new AppError("FORBIDDEN", "Only the client can abandon this contract.", 403);
    if (contract.status !== "funding_pending") throw new AppError("CONTRACT_FUNDED", "Only an unfunded contract can be abandoned. Use cancel for a funded escrow.", 409);
    const signed = await conn.query("select 1 from chain_actions a join chain_signatures s on s.chain_action_id = a.id where a.contract_id = $1 limit 1", [contractId]);
    if (signed.rowCount) throw new AppError("FUNDING_IN_FLIGHT", "A funding transaction was already submitted. Wait for it to finalize or expire.", 409);
    await conn.query("delete from chain_actions where contract_id = $1", [contractId]);
    await conn.query("delete from contracts where id = $1", [contractId]);
    await conn.query("commit");
  } catch (error) {
    await conn.query("rollback");
    throw error;
  } finally {
    conn.release();
  }
  return { id: contractId, abandoned: true };
}
