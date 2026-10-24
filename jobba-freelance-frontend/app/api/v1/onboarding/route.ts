import { saveOnboardingSchema } from "@/features/profiles/schemas";
import { requireActor } from "@/server/services/auth-context";
import { handleRoute, mutationGuard, ok, parseJson } from "@/server/services/http";
import { readOnboarding, saveOnboarding } from "@/server/onboarding/service";

export async function GET() {
  return handleRoute(async () => ok(await readOnboarding((await requireActor(undefined, { allowIncomplete: true })).profileId)));
}

export async function PATCH(request: Request) {
  return handleRoute(async () => {
    const originError = mutationGuard(request);
    if (originError) return originError;
    const actor = await requireActor(undefined, { allowIncomplete: true });
    return ok(await saveOnboarding(actor, await parseJson(request, saveOnboardingSchema)));
  });
}
