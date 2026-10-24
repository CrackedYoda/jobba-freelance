import { SiteHeader } from "@/components/marketplace/site-header";
import { ProfileEditor } from "@/features/profiles/profile-editor";

export default function OnboardingPage() {
  return <><SiteHeader /><main className="shell narrowPage pageSection"><div className="onboardingProgress"><span>Profile setup</span><strong>1 of 1</strong><i><b /></i></div><div className="pageIntro"><p className="eyebrow">Your profile</p><h1>Show people how you <em>work best.</em></h1><p className="lead">Choose how you use Jobba and keep your public details current.</p></div><ProfileEditor /></main></>;
}
