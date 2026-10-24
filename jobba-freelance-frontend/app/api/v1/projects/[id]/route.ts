import { projectIdSchema, updateProjectSchema } from "@/features/projects/schemas";
import { requireActor } from "@/server/services/auth-context";
import { handleRoute, mutationGuard, ok, parseJson } from "@/server/services/http";
import { readProject, updateProject } from "@/server/services/projects";

type Context = { params: Promise<{ id: string }> };

export async function GET(_request: Request, context: Context) {
  return handleRoute(async () => {
    const projectId = projectIdSchema.parse((await context.params).id);
    return ok(await readProject(projectId));
  });
}

export async function PATCH(request: Request, context: Context) {
  return handleRoute(async () => {
    const originError = mutationGuard(request);
    if (originError) return originError;
    const actor = await requireActor("client");
    const projectId = projectIdSchema.parse((await context.params).id);
    return ok(await updateProject(actor, projectId, await parseJson(request, updateProjectSchema)));
  });
}
