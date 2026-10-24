import { z } from "zod";

import { requireActor } from "@/server/services/auth-context";
import { handleRoute, mutationGuard, ok } from "@/server/services/http";
import { abandonUnfundedContract } from "@/server/services/proposals";

type Context = { params: Promise<{ id: string }> };

export async function POST(request: Request, context: Context) {
  return handleRoute(async () => {
    const originError = mutationGuard(request);
    if (originError) return originError;
    const actor = await requireActor("client");
    const contractId = z.string().uuid().parse((await context.params).id);
    return ok(await abandonUnfundedContract(actor, contractId));
  });
}
