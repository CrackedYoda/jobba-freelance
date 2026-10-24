import "server-only";

import type { CreateProjectInput, ListProjectsInput, UpdateProjectInput } from "@/features/projects/schemas";
import { getServerEnv } from "@/lib/config/env";
import { decodeCursor, encodeCursor } from "@/server/db/cursor";
import { getDatabase } from "@/server/db/client";
import { AppError } from "@/server/db/errors";
import { getProjectById, insertProject, listProjects } from "@/server/repositories/projects";
import type { Actor } from "@/server/repositories/profiles";
import { serializeProject } from "@/server/services/serialize";

export async function createProject(actor: Actor, input: CreateProjectInput) {
  if (!actor.roles.includes("client")) throw new AppError("FORBIDDEN", "The client role is required.", 403);
  const mint = getServerEnv().SOLANA_USDC_MINT;
  if (!mint) throw new AppError("SERVER_NOT_CONFIGURED", "The USDC mint is not configured.", 503);
  const project = await insertProject(getDatabase(), actor.profileId, mint, input);
  if (!project) throw new AppError("DATABASE_ERROR", "The project could not be loaded after creation.", 500);
  return serializeProject(project as Record<string, unknown>);
}

export async function browseProjects(input: ListProjectsInput) {
  const rows = await listProjects(getDatabase(), input, decodeCursor(input.cursor));
  const hasMore = rows.length > input.limit;
  const page = rows.slice(0, input.limit);
  const last = page.at(-1) as Record<string, unknown> | undefined;
  return {
    items: page.map((row) => serializeProject(row as Record<string, unknown>)),
    nextCursor: hasMore && last ? encodeCursor({ createdAt: String(last.created_at), id: String(last.id) }) : null,
  };
}

export async function readProject(projectId: string) {
  const project = await getProjectById(getDatabase(), projectId);
  if (!project || project.status === "draft") throw new AppError("PROJECT_NOT_FOUND", "Project not found.", 404);
  return serializeProject(project as Record<string, unknown>);
}

const projectTransitions: Record<string, readonly string[]> = { draft: ["open", "cancelled"], open: ["cancelled"] };

export async function updateProject(actor: Actor, projectId: string, input: UpdateProjectInput) {
  const db = getDatabase();
  const conn = await db.connect();
  try {
    await conn.query("begin");
    const { rows } = await conn.query<{ client_id: string; status: string }>("select client_id, status from projects where id = $1 for update", [projectId]);
    const current = rows[0];
    if (!current) throw new AppError("PROJECT_NOT_FOUND", "Project not found.", 404);
    if (current.client_id !== actor.profileId) throw new AppError("FORBIDDEN", "Only the project client can change it.", 403);
    if (!["draft", "open"].includes(current.status)) throw new AppError("PROJECT_LOCKED", `A ${current.status} project can no longer be changed.`, 409);
    if (input.status && input.status !== current.status && !projectTransitions[current.status]?.includes(input.status)) {
      throw new AppError("INVALID_PROJECT_TRANSITION", `A ${current.status} project cannot become ${input.status}.`, 409);
    }
    if (input.status === "cancelled") {
      const contract = await conn.query("select 1 from contracts where project_id = $1", [projectId]);
      if (contract.rowCount) throw new AppError("PROJECT_HAS_CONTRACT", "Cancel or abandon the contract before cancelling the project.", 409);
    }
    const editsTerms = input.budgetAtomic !== undefined || input.category !== undefined;
    if (editsTerms && current.status === "open") {
      const proposals = await conn.query("select 1 from proposals where project_id = $1 and status = 'pending' limit 1", [projectId]);
      if (proposals.rowCount) throw new AppError("PROJECT_HAS_PROPOSALS", "Budget and category are fixed once proposals arrive.", 409);
    }
    await conn.query(
      `update projects set title = coalesce($2, title), description = coalesce($3, description), category = coalesce($4, category),
              budget_atomic = coalesce($5::bigint, budget_atomic), delivery_days = coalesce($6, delivery_days), status = coalesce($7::project_status, status),
              published_at = case when $7 = 'open' and published_at is null then now() else published_at end, version = version + 1
        where id = $1`,
      [projectId, input.title ?? null, input.description ?? null, input.category ?? null, input.budgetAtomic ?? null, input.deliveryDays ?? null, input.status ?? null],
    );
    if (input.skillIds) {
      await conn.query("delete from project_skills where project_id = $1", [projectId]);
      if (input.skillIds.length) await conn.query("insert into project_skills (project_id, skill_id) select $1, unnest($2::uuid[])", [projectId, input.skillIds]);
    }
    if (input.status === "cancelled") {
      await conn.query("update proposals set status = 'rejected', version = version + 1 where project_id = $1 and status = 'pending'", [projectId]);
    }
    await conn.query("commit");
  } catch (error) {
    await conn.query("rollback");
    throw error;
  } finally {
    conn.release();
  }
  const project = await getProjectById(db, projectId);
  return serializeProject(project as Record<string, unknown>);
}

export async function readOwnProject(actor: Actor, projectId: string) {
  const project = await getProjectById(getDatabase(), projectId);
  if (!project || (project.status === "draft" && project.client_id !== actor.profileId)) throw new AppError("PROJECT_NOT_FOUND", "Project not found.", 404);
  return serializeProject(project as Record<string, unknown>);
}
