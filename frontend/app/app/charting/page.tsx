import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";
import { Charting } from "@/components/charting";

export default async function ChartingPage() {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  return (
    <div className="flex-1 w-full flex flex-col gap-12">
      <Charting />
    </div>
  );
}

