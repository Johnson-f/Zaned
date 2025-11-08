"use client"

import * as React from "react"
import { Plus, Star } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useWatchlists, useAddWatchlistItem } from "@/hooks/use-watchlist"

const BUILT_IN_WATCHLIST_NAMES = [
  "Inside Day",
  "Highest Volume Quarter",
  "Highest Volume Year",
  "Highest Volume Ever",
]

interface AddToWatchlistDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  symbols: string[]
  builtInWatchlistName: string
}

export function AddToWatchlistDialog({
  open,
  onOpenChange,
  symbols,
  builtInWatchlistName,
}: AddToWatchlistDialogProps) {
  const { data: watchlists = [] } = useWatchlists(true)
  const addWatchlistItem = useAddWatchlistItem()
  const [selectedWatchlistId, setSelectedWatchlistId] = React.useState<string>("")
  const [addingSymbol, setAddingSymbol] = React.useState<string | null>(null)
  const [openMiniModal, setOpenMiniModal] = React.useState(false)
  const [selectedSymbolForModal, setSelectedSymbolForModal] = React.useState<string | null>(null)

  // Filter out built-in watchlists to only show user-created ones
  const userWatchlists = React.useMemo(() => {
    return watchlists.filter(
      (w) => !BUILT_IN_WATCHLIST_NAMES.some(
        (name) => w.name.toLowerCase() === name.toLowerCase()
      )
    )
  }, [watchlists])

  // Reset selected watchlist when dialog opens/closes
  React.useEffect(() => {
    if (!open) {
      setSelectedWatchlistId("")
      setAddingSymbol(null)
      setOpenMiniModal(false)
      setSelectedSymbolForModal(null)
    }
  }, [open])

  const handleStarClick = (symbol: string) => {
    setSelectedSymbolForModal(symbol)
    setOpenMiniModal(true)
  }

  const handleAddToWatchlist = async (watchlistId: string, symbol: string) => {
    setAddingSymbol(symbol)
    try {
      await addWatchlistItem.mutateAsync({
        watchlistId,
        item: { symbol, name: symbol },
      })
      setOpenMiniModal(false)
      setSelectedSymbolForModal(null)
    } catch (error) {
      console.error("Failed to add stock to watchlist:", error)
    } finally {
      setAddingSymbol(null)
    }
  }

  const handleAddAll = async () => {
    if (!selectedWatchlistId || symbols.length === 0) return

    for (const symbol of symbols) {
      try {
        await addWatchlistItem.mutateAsync({
          watchlistId: selectedWatchlistId,
          item: { symbol, name: symbol },
        })
      } catch (error) {
        console.error(`Failed to add ${symbol} to watchlist:`, error)
      }
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Add Stocks to Watchlist</DialogTitle>
          <DialogDescription>
            Select a watchlist and add stocks from {builtInWatchlistName}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 flex-1 overflow-hidden flex flex-col">
          {/* Watchlist Selector */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Select Watchlist</label>
            <Select value={selectedWatchlistId} onValueChange={setSelectedWatchlistId}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Choose a watchlist..." />
              </SelectTrigger>
              <SelectContent>
                {userWatchlists.length === 0 ? (
                  <div className="px-2 py-1.5 text-sm text-muted-foreground">
                    No watchlists available. Create one first.
                  </div>
                ) : (
                  userWatchlists.map((watchlist) => (
                    <SelectItem key={watchlist.id} value={watchlist.id}>
                      {watchlist.name} ({watchlist.items?.length || 0})
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          </div>

          {/* Add All Button */}
          {selectedWatchlistId && symbols.length > 0 && (
            <div className="flex justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={handleAddAll}
                disabled={addWatchlistItem.isPending}
                className="gap-2"
              >
                <Plus className="size-4" />
                Add All ({symbols.length})
              </Button>
            </div>
          )}

          {/* Stocks List */}
          <div className="flex-1 min-h-0 border rounded-md">
            <ScrollArea className="h-full">
              <div className="p-2 space-y-1">
                {symbols.length === 0 ? (
                  <div className="text-sm text-muted-foreground text-center py-4">
                    No stocks available
                  </div>
                ) : (
                  symbols.map((symbol) => (
                    <div
                      key={symbol}
                      className="flex items-center justify-between gap-3 px-3 py-2 rounded-md hover:bg-accent transition-colors"
                    >
                      <div className="flex-1 min-w-0">
                        <div className="font-medium text-sm">{symbol}</div>
                      </div>
                      <button
                        onClick={() => handleStarClick(symbol)}
                        disabled={addingSymbol === symbol || addWatchlistItem.isPending}
                        className="p-1.5 rounded-md hover:bg-accent transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        title="Add to watchlist"
                      >
                        <Star className="size-4 text-muted-foreground hover:text-yellow-500 transition-colors" />
                      </button>
                    </div>
                  ))
                )}
              </div>
            </ScrollArea>
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-4 border-t">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            Close
          </Button>
        </div>
      </DialogContent>

      {/* Mini Modal for Selecting Watchlist */}
      <Dialog open={openMiniModal} onOpenChange={setOpenMiniModal}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Add {selectedSymbolForModal} to Watchlist</DialogTitle>
            <DialogDescription>
              Select a watchlist to add this stock
            </DialogDescription>
          </DialogHeader>

          <ScrollArea className="max-h-[300px]">
            <div className="space-y-1 pr-4">
              {userWatchlists.length === 0 ? (
                <div className="text-sm text-muted-foreground text-center py-4">
                  No watchlists available. Create one first.
                </div>
              ) : (
                userWatchlists.map((watchlist) => (
                  <button
                    key={watchlist.id}
                    onClick={() => {
                      if (selectedSymbolForModal) {
                        handleAddToWatchlist(watchlist.id, selectedSymbolForModal)
                      }
                    }}
                    disabled={addingSymbol === selectedSymbolForModal || addWatchlistItem.isPending}
                    className="w-full px-3 py-2 text-left rounded-md hover:bg-accent transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-sm">{watchlist.name}</span>
                      <span className="text-xs text-muted-foreground">
                        ({watchlist.items?.length || 0})
                      </span>
                    </div>
                  </button>
                ))
              )}
            </div>
          </ScrollArea>

          <div className="flex justify-end gap-2 pt-4 border-t">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setOpenMiniModal(false)}
            >
              Cancel
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </Dialog>
  )
}

