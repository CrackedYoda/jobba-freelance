import { z } from "zod";

import { readAvatar } from "@/server/onboarding/service";

export async function GET(_request: Request, context: { params: Promise<{ id: string }> }) {
  const id = z.string().uuid().safeParse((await context.params).id);
  const avatar = id.success ? await readAvatar(id.data) : undefined;
  if (!avatar) return new Response("Not found", { status: 404 });
  return new Response(new Uint8Array(avatar.bytes), {
    headers: { "Content-Type": avatar.content_type, "Cache-Control": "public, max-age=31536000, immutable", "X-Content-Type-Options": "nosniff", "Content-Security-Policy": "default-src 'none'" },
  });
}
