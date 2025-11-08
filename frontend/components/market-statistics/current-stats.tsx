"use client"

import * as React from "react"
import { useLiveMarketStatistics } from "@/hooks/use-market-statistics"
import { Loader2, TrendingUp, TrendingDown, Minus } from "lucide-react"

export function CurrentStats() {
  const { data: stats, isLoading, error } = useLiveMarketStatistics(true)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-32">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-sm text-red-400 text-center py-4">
        Error loading market statistics: {error instanceof Error ? error.message : "Unknown error"}
      </div>
    )
  }

  if (!stats) {
    return (
      <div className="text-sm text-muted-foreground text-center py-4">
        No market statistics available
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {/* Advancers */}
        <div className="bg-card border rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-muted-foreground">Advancers</span>
            <TrendingUp className="h-4 w-4 text-green-500" />
          </div>
          <div className="text-2xl font-bold text-green-500">{stats.advances.toLocaleString()}</div>
        </div>

        {/* Decliners */}
        <div className="bg-card border rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-muted-foreground">Decliners</span>
            <TrendingDown className="h-4 w-4 text-red-500" />
          </div>
          <div className="text-2xl font-bold text-red-500">{stats.decliners.toLocaleString()}</div>
        </div>

        {/* Unchanged */}
        <div className="bg-card border rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-muted-foreground">Unchanged</span>
            <Minus className="h-4 w-4 text-gray-500" />
          </div>
          <div className="text-2xl font-bold text-gray-500">{stats.unchanged.toLocaleString()}</div>
        </div>

        {/* Total */}
        <div className="bg-card border rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-muted-foreground">Total</span>
          </div>
          <div className="text-2xl font-bold">{stats.total.toLocaleString()}</div>
        </div>
      </div>

      {/* Last Updated Timestamp */}
      {stats.last_updated && (
        <div className="text-xs text-muted-foreground text-center">
          Last updated: {new Date(stats.last_updated).toLocaleString()}
        </div>
      )}
    </div>
  )
}

