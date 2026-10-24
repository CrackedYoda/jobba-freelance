import { requireActor } from "@/server/services/auth-context";
import { handleRoute, mutationGuard, ok } from "@/server/services/http";
import { verifyBalance } from "@/server/onboarding/service";

export async function POST(request: Request) {
  return handleRoute(async () => {
    const originError = mutationGuard(request);
    if (originError) return originError;
    return ok(await verifyBalance(await requireActor(undefined, { allowIncomplete: true })));
  });
}
