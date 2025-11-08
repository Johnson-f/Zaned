"use client"

import * as React from "react"
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts"
import { useMarketStatistics } from "@/hooks/use-market-statistics"
import type { MarketStatistics } from "@/lib/types/market-statistics"
import { Loader2 } from "lucide-react"

interface MarketStatisticsChartProps {
  startDate?: string
  endDate?: string
}

export function MarketStatisticsChart({ startDate, endDate }: MarketStatisticsChartProps) {
  // Fetch historical market statistics
  const { data: statisticsData = [], isLoading, error } = useMarketStatistics(
    startDate,
    endDate,
    true
  )

  // Transform data for recharts format
  const chartData = React.useMemo(() => {
    return statisticsData.map((stat: MarketStatistics) => ({
      date: new Date(stat.date).toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
      }),
      fullDate: stat.date,
      Advancers: stat.stocksUp,
      Decliners: stat.stocksDown,
      Unchanged: stat.stocksUnchanged,
    }))
  }, [statisticsData])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-[500px] bg-[#0a0a0a]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-[500px] bg-[#0a0a0a] text-red-400">
        Error loading chart data: {error instanceof Error ? error.message : "Unknown error"}
      </div>
    )
  }

  if (!statisticsData.length) {
    return (
      <div className="flex items-center justify-center h-[500px] bg-[#0a0a0a] text-muted-foreground">
        No market statistics data available
      </div>
    )
  }

  return (
    <div className="w-full bg-[#0a0a0a] p-4">
      <ResponsiveContainer width="100%" height={500}>
        <LineChart
          data={chartData}
          margin={{
            top: 5,
            right: 30,
            left: 20,
            bottom: 5,
          }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" />
          <XAxis
            dataKey="date"
            stroke="#d1d5db"
            style={{ fontSize: "12px" }}
            tick={{ fill: "#d1d5db" }}
          />
          <YAxis
            stroke="#d1d5db"
            style={{ fontSize: "12px" }}
            tick={{ fill: "#d1d5db" }}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: "#1f2937",
              border: "1px solid #374151",
              borderRadius: "6px",
              color: "#d1d5db",
            }}
            labelStyle={{ color: "#d1d5db" }}
          />
          <Legend
            wrapperStyle={{ color: "#d1d5db" }}
            iconType="line"
          />
          <Line
            type="monotone"
            dataKey="Advancers"
            stroke="#10b981"
            strokeWidth={2}
            dot={false}
            name="Advancers"
          />
          <Line
            type="monotone"
            dataKey="Decliners"
            stroke="#ef4444"
            strokeWidth={2}
            dot={false}
            name="Decliners"
          />
          <Line
            type="monotone"
            dataKey="Unchanged"
            stroke="#6b7280"
            strokeWidth={2}
            dot={false}
            name="Unchanged"
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

