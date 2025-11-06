"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Star } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useHighVolumeQuarterSymbols } from "@/hooks/use-historical"
import { useWatchlists, useAddWatchlistItem } from "@/hooks/use-watchlist"

const BUILT_IN_WATCHLIST_NAME = "Highest Volume Quarter"

export default function HighestVolumeQuarterWatchlist() {
  const router = useRouter()
  const { data: highVolumeQuarterData, isLoading, error } = useHighVolumeQuarterSymbols(true)
  const symbols: string[] = highVolumeQuarterData?.symbols || []

  const { data: watchlists = [] } = useWatchlists(true)
  const addWatchlistItem = useAddWatchlistItem()

  // Find the "Highest Volume Quarter" watchlist ID
  const builtInWatchlistId = React.useMemo(() => {
    const existing = watchlists.find((w) => w.name.toLowerCase() === BUILT_IN_WATCHLIST_NAME.toLowerCase())
    return existing?.id || ""
  }, [watchlists])

  const handleAdd = async (symbol: string) => {
    if (!builtInWatchlistId) return
    try {
      await addWatchlistItem.mutateAsync({
        watchlistId: builtInWatchlistId,
        item: { symbol, name: symbol },
      })
    } catch {
      // ignore
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">Loading...</div>
        ) : error ? (
          <div className="flex items-center justify-center py-8 text-sm text-destructive">Failed to load</div>
        ) : symbols.length === 0 ? (
          <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">No matches found</div>
        ) : (
          <div className="divide-y">
            {symbols.map((sym) => (
              <div
                key={sym}
                className="px-4 py-3 hover:bg-sidebar-accent/50 transition-colors cursor-pointer select-none"
                onDoubleClick={() => router.push(`/app/charting?symbol=${sym}`)}
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-sm whitespace-nowrap">{sym}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={(e) => {
                        e.stopPropagation()
                        void handleAdd(sym)
                      }}
                      disabled={!builtInWatchlistId || addWatchlistItem.isPending}
                      title="Add to Highest Volume Quarter watchlist"
                    >
                      <Star className="size-4 text-muted-foreground hover:text-yellow-500" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation()
                        router.push(`/app/charting?symbol=${sym}`)
                      }}
                    >
                      View
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

