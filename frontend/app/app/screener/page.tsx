import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";
import { Screener } from "@/components/screener";

export default async function ScreenerPage() {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  return (
    <div className="flex-1 w-full flex flex-col gap-12">
      <Screener />
    </div>
  );
}

