"use client"

import { Suspense } from "react"
import { useSearchParams } from "next/navigation"
import { Star } from "lucide-react"

function WatchlistContent() {
  const searchParams = useSearchParams()
  const symbol = searchParams.get("symbol")

  if (symbol) {
    // TODO: Show symbol details when symbol is selected
    return (
      <div className="flex-1 w-full flex flex-col">
        {/* Symbol details will go here */}
      </div>
    )
  }

  return (
    <div className="flex-1 w-full flex flex-col items-center justify-center">
      <div className="flex flex-col items-center gap-4 text-center max-w-md px-4">
        <div className="rounded-full bg-muted p-6">
          <Star className="size-12 text-muted-foreground" />
        </div>
        <div className="space-y-2">
          <h3 className="text-lg font-semibold">No symbol selected</h3>
          <p className="text-sm text-muted-foreground">
            Select a symbol from your watchlist to view details and charts.
          </p>
        </div>
      </div>
    </div>
  )
}

export default function WatchlistPage() {
  return (
    <Suspense fallback={
      <div className="flex-1 w-full flex flex-col items-center justify-center">
        <div className="text-sm text-muted-foreground">Loading...</div>
      </div>
    }>
      <WatchlistContent />
    </Suspense>
  )
}

