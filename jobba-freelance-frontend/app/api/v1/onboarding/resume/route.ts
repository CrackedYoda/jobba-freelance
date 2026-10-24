import { NextResponse } from "next/server";

import { takeRateLimit } from "@/server/auth/rate-limit";
import { AppError } from "@/server/db/errors";
import { requireActor } from "@/server/services/auth-context";
import { handleRoute, mutationGuard, ok } from "@/server/services/http";
import { importResume, readOwnResume } from "@/server/onboarding/resume";

/** Upload a résumé and get back a draft profile for the freelancer to review. */
export async function POST(request: Request) {
  return handleRoute(async () => {
    const originError = mutationGuard(request);
    if (originError) return originError;
    const actor = await requireActor(undefined, { allowIncomplete: true });
    if (!takeRateLimit(`resume:${actor.profileId}`, 5, 60 * 60_000)) {
      throw new AppError("RATE_LIMITED", "You've imported several résumés recently. Please try again later.", 429);
    }
    const file = (await request.formData().catch(() => null))?.get("file");
    if (!(file instanceof File)) throw new AppError("INVALID_RESUME", "Attach a PDF or Word (.docx) file.", 400);
    return ok(await importResume(actor, file));
  });
}

/** Download your own stored résumé. Résumés are never served to anyone else. */
export async function GET() {
  return handleRoute(async () => {
    const actor = await requireActor(undefined, { allowIncomplete: true });
    const resume = await readOwnResume(actor.profileId);
    if (!resume) throw new AppError("RESUME_NOT_FOUND", "No résumé uploaded yet.", 404);
    return new NextResponse(new Uint8Array(resume.bytes), {
      headers: {
        "Content-Type": resume.content_type,
        "Content-Disposition": `attachment; filename="${resume.file_name.replace(/["\\\r\n]/g, "_")}"`,
        "Cache-Control": "private, no-store",
        "X-Content-Type-Options": "nosniff",
      },
    });
  });
}
