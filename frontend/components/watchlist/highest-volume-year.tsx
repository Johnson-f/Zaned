"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Plus, Star } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useHighVolumeYearSymbols } from "@/hooks/use-historical"
import { useWatchlists, useAddWatchlistItem } from "@/hooks/use-watchlist"
import { AddToWatchlistDialog } from "@/components/watchlist/AddToWatchlistDialog"
import { BuiltInStockActionMenu } from "@/components/watchlist/BuiltInStockActionMenu"

const BUILT_IN_WATCHLIST_NAME = "Highest Volume Year"

export default function HighestVolumeYearWatchlist() {
  const router = useRouter()
  const { data: highVolumeYearData, isLoading, error } = useHighVolumeYearSymbols(true)
  const symbols: string[] = highVolumeYearData?.symbols || []
  const [openAddDialog, setOpenAddDialog] = React.useState(false)
  const [openItemMenu, setOpenItemMenu] = React.useState(false)
  const [menuPosition, setMenuPosition] = React.useState<{ x: number; y: number } | null>(null)
  const [selectedItem, setSelectedItem] = React.useState<{ symbol: string } | null>(null)

  const { data: watchlists = [] } = useWatchlists(true)
  const addWatchlistItem = useAddWatchlistItem()

  // Find the "Highest Volume Year" watchlist ID
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

  // Handle double-click or right-click on stock item
  const handleItemAction = (e: React.MouseEvent, symbol: string) => {
    e.preventDefault()
    e.stopPropagation()
    setSelectedItem({ symbol })
    setMenuPosition({ x: e.clientX, y: e.clientY })
    setOpenItemMenu(true)
  }

  const handleItemDoubleClick = (e: React.MouseEvent, symbol: string) => {
    handleItemAction(e, symbol)
  }

  const handleItemRightClick = (e: React.MouseEvent, symbol: string) => {
    e.preventDefault()
    e.stopPropagation()
    handleItemAction(e, symbol)
  }

  return (
    <div className="flex flex-col h-full">
      {/* Add Stocks Button */}
      {symbols.length > 0 && (
        <div className="px-4 py-2 border-b">
          <Button
            variant="outline"
            size="sm"
            className="gap-2"
            onClick={() => setOpenAddDialog(true)}
          >
            <Plus className="size-4" />
            Add stocks
          </Button>
        </div>
      )}

      {/* Add To Watchlist Dialog */}
      <AddToWatchlistDialog
        open={openAddDialog}
        onOpenChange={setOpenAddDialog}
        symbols={symbols}
        builtInWatchlistName={BUILT_IN_WATCHLIST_NAME}
      />

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
                onDoubleClick={(e) => handleItemDoubleClick(e, sym)}
                onContextMenu={(e) => handleItemRightClick(e, sym)}
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
                      title="Add to Highest Volume Year watchlist"
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

      {/* Context Menu */}
      <BuiltInStockActionMenu
        open={openItemMenu}
        position={menuPosition}
        selectedItem={selectedItem}
        onClose={() => {
          setOpenItemMenu(false)
          setSelectedItem(null)
        }}
        onAddToWatchlist={() => setOpenAddDialog(true)}
      />
    </div>
  )
}

