import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";
import { HistoricalTest } from "@/components/historical-test";

export default async function WatchlistPage() {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  return (
    <div className="flex-1 w-full flex flex-col gap-12">
      <div className="flex flex-col gap-2 items-start">
        <h2 className="font-bold text-2xl mb-4">Watchlist</h2>
        <p className="text-muted-foreground mb-6">
          Test component using TanStack Query hooks to fetch historical data from the backend.
        </p>
      </div>
      <HistoricalTest />
    </div>
  );
}

