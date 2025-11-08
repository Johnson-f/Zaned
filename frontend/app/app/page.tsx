import { redirect } from "next/navigation";

import { createClient } from "@/lib/supabase/server";
import { CurrentStats } from "@/components/market-statistics/current-stats";
import { MarketStatisticsChart } from "@/components/market-statistics/market-statistics-chart";

export default async function ProtectedPage() {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  // Calculate date range for chart (last 30 days)
  const endDate = new Date();
  const startDate = new Date();
  startDate.setDate(startDate.getDate() - 30);
  
  const startDateStr = startDate.toISOString().split("T")[0];
  const endDateStr = endDate.toISOString().split("T")[0];

  return (
    <div className="flex-1 w-full flex flex-col gap-8">
      {/* Market Statistics Section */}
      <div className="flex flex-col gap-6">
        <div>
          <p className="text-muted-foreground text-sm">
            Real-time market statistics showing advancers, decliners, and unchanged stocks
          </p>
        </div>
        
        {/* Current Stats Cards */}
        <div className="w-full">
          <h3 className="font-semibold text-lg mb-4">Today&apos;s Statistics</h3>
          <CurrentStats />
        </div>
        
        {/* Historical Chart */}
        <div className="w-full">
          <h3 className="font-semibold text-lg mb-4">Historical Trends (Last 30 Days)</h3>
          <MarketStatisticsChart startDate={startDateStr} endDate={endDateStr} />
        </div>
      </div>
    </div>
  );
}
