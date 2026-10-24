import type { Metadata } from "next";

import { ApiForm } from "@/components/marketplace/api-form";
import { SiteHeader } from "@/components/marketplace/site-header";

export const metadata: Metadata = { title: "Post a project | Jobba" };

export default function NewProjectPage() {
  return (
    <><SiteHeader /><main className="shell narrowPage pageSection"><div className="pageIntro"><p className="eyebrow">New project</p><h1>Give great work a <em>clear start.</em></h1><p className="lead">Set the outcome, fixed USDC budget, and delivery expectations freelancers will propose against.</p></div>
      <ApiForm endpoint="/api/v1/projects" kind="project">
        <fieldset><legend>Project basics</legend><label><span>Project title</span><input name="title" required minLength={8} maxLength={120} placeholder="e.g. Build an Anchor event indexer" /></label><label><span>Category</span><select name="category" required defaultValue=""><option value="" disabled>Select a category</option><option>Development</option><option>Design & Creative</option><option>Writing</option><option>Marketing</option></select></label><label className="fullField"><span>What needs to be delivered?</span><textarea name="description" required minLength={40} maxLength={4000} rows={7} placeholder="Describe the outcome, important context, and what a strong delivery includes." /><small>Do not include private credentials or sensitive customer data.</small></label></fieldset>
        <fieldset><legend>Terms</legend><label><span>Fixed budget</span><div className="inputSuffix"><input name="budget" required inputMode="decimal" pattern="\d+(\.\d{1,6})?" placeholder="2400" /><b>USDC</b></div></label><label><span>Delivery target</span><div className="inputSuffix"><input name="deliveryDays" type="number" required min={1} max={365} placeholder="21" /><b>days</b></div></label><label className="fullField"><span>Required skills</span><input name="skills" required placeholder="Product design, Figma, design systems" /><small>Separate skills with commas.</small></label></fieldset>
        <div className="formActions"><button className="button ghost" type="submit" name="intent" value="draft">Save draft</button><button className="button primary" type="submit" name="intent" value="open">Publish project</button></div>
      </ApiForm>
    </main></>
  );
}
