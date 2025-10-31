import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";

export default async function ScreenerPage() {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  return (
    <div className="flex-1 w-full flex flex-col gap-12">
      <div className="flex flex-col gap-2 items-start">
        <h2 className="font-bold text-2xl mb-4">Screener</h2>
        <p className="text-muted-foreground">
          Your screener page content will go here.
        </p>
      </div>
    </div>
  );
}

