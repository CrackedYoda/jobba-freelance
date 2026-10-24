import { AppError } from "@/server/db/errors";
import { requireActor } from "@/server/services/auth-context";
import { handleRoute, mutationGuard, ok } from "@/server/services/http";
import { saveAvatar } from "@/server/onboarding/service";

export async function POST(request: Request) {
  return handleRoute(async () => {
    const originError = mutationGuard(request);
    if (originError) return originError;
    const actor = await requireActor(undefined, { allowIncomplete: true });
    const file = (await request.formData().catch(() => null))?.get("file");
    if (!(file instanceof File)) throw new AppError("INVALID_IMAGE", "Attach an image file.", 400);
    return ok(await saveAvatar(actor, file));
  });
}
