"use client"

import * as React from "react"
import { Plus } from "lucide-react"
import { Select, SelectContent, SelectItem, SelectSeparator, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import type { Watchlist } from "@/lib/types/watchlist"

export function WatchlistSelect({
  value,
  onValueChange,
  watchlists,
  loading,
  selectedWatchlist,
  itemCount,
  insideDayCount,
  highestVolumeQuarterCount,
  highestVolumeYearCount,
  highestVolumeEverCount,
  onCreateWatchlist,
  isCreating,
}: {
  value: string | undefined
  onValueChange: (val: string) => void
  watchlists: Watchlist[]
  loading: boolean
  selectedWatchlist?: Watchlist | null
  itemCount: number
  insideDayCount?: number
  highestVolumeQuarterCount?: number
  highestVolumeYearCount?: number
  highestVolumeEverCount?: number
  onCreateWatchlist?: (name: string) => Promise<void>
  isCreating?: boolean
}) {
  const isInsideDaySelected = value === "inside-day"
  const isHighestVolumeQuarterSelected = value === "highest-volume-quarter"
  const isHighestVolumeYearSelected = value === "highest-volume-year"
  const isHighestVolumeEverSelected = value === "highest-volume-ever"
  const [openCreatePopover, setOpenCreatePopover] = React.useState(false)
  const [watchlistName, setWatchlistName] = React.useState("")
  
  // Filter out "Inside Day", "Highest Volume Quarter", "Highest Volume Year", and "Highest Volume Ever" watchlists to prevent duplicates since we have special options
  const filteredWatchlists = watchlists.filter(
    (w) => w.name.toLowerCase() !== "inside day" && 
           w.name.toLowerCase() !== "highest volume quarter" &&
           w.name.toLowerCase() !== "highest volume year" &&
           w.name.toLowerCase() !== "highest volume ever"
  )
  
  const handleCreateWatchlist = async () => {
    if (!watchlistName.trim() || !onCreateWatchlist) return
    
    try {
      await onCreateWatchlist(watchlistName.trim())
      setWatchlistName("")
      setOpenCreatePopover(false)
    } catch (error) {
      console.error("Failed to create watchlist:", error)
    }
  }
  
  return (
    <div className="flex items-center gap-2">
      <Select
        value={value}
        onValueChange={onValueChange}
        disabled={loading}
      >
        <SelectTrigger className="w-full">
          <SelectValue>
            {loading
              ? "Loading..."
              : isInsideDaySelected
              ? `Inside (${insideDayCount ?? 0})`
              : isHighestVolumeQuarterSelected
              ? `Highest Volume Quarter (HVQ) (${highestVolumeQuarterCount ?? 0})`
              : isHighestVolumeYearSelected
              ? `Highest Volume Year (HVY) (${highestVolumeYearCount ?? 0})`
              : isHighestVolumeEverSelected
              ? `Highest Volume Ever (HVE) (${highestVolumeEverCount ?? 0})`
              : filteredWatchlists.length === 0
              ? "Select watchlist"
              : selectedWatchlist
              ? `${selectedWatchlist.name} (${itemCount})`
              : "Select watchlist"}
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          {insideDayCount !== undefined && (
            <SelectItem value="inside-day">
              Inside ({insideDayCount})
            </SelectItem>
          )}
          {highestVolumeQuarterCount !== undefined && (
            <SelectItem value="highest-volume-quarter">
              Highest Volume Q ({highestVolumeQuarterCount})
            </SelectItem>
          )}
          {highestVolumeYearCount !== undefined && (
            <SelectItem value="highest-volume-year">
              Highest Volume Y ({highestVolumeYearCount})
            </SelectItem>
          )}
          {highestVolumeEverCount !== undefined && (
            <SelectItem value="highest-volume-ever">
              Highest Volume E ({highestVolumeEverCount})
            </SelectItem>
          )}
          {filteredWatchlists.length === 0 ? (
            <div className="px-2 py-1.5 text-sm text-muted-foreground">
              No watchlists yet
            </div>
          ) : (
            filteredWatchlists.map((watchlist) => (
              <SelectItem key={watchlist.id} value={watchlist.id}>
                {watchlist.name} ({watchlist.items?.length || 0})
              </SelectItem>
            ))
          )}
          {onCreateWatchlist && (
            <>
              <SelectSeparator />
              <Popover open={openCreatePopover} onOpenChange={setOpenCreatePopover}>
                <PopoverTrigger asChild>
                  <div className="px-2 py-1.5">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start gap-2 h-auto py-1.5"
                      onClick={(e) => {
                        e.preventDefault()
                        e.stopPropagation()
                        setOpenCreatePopover(true)
                      }}
                    >
                      <Plus className="size-4" />
                      <span className="text-sm">Create New Watchlist</span>
                    </Button>
                  </div>
                </PopoverTrigger>
                <PopoverContent className="w-80" align="start" onClick={(e) => e.stopPropagation()}>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <h4 className="font-medium text-sm">Create Watchlist</h4>
                      <p className="text-xs text-muted-foreground">
                        Enter a name for your new watchlist
                      </p>
                    </div>
                    <div className="space-y-2">
                      <Input
                        placeholder="Watchlist name"
                        value={watchlistName}
                        onChange={(e) => setWatchlistName(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") {
                            void handleCreateWatchlist()
                          }
                        }}
                        autoFocus
                      />
                    </div>
                    <div className="flex justify-end gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setOpenCreatePopover(false)
                          setWatchlistName("")
                        }}
                      >
                        Cancel
                      </Button>
                      <Button
                        size="sm"
                        onClick={handleCreateWatchlist}
                        disabled={!watchlistName.trim() || isCreating}
                      >
                        {isCreating ? "Creating..." : "Create"}
                      </Button>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            </>
          )}
        </SelectContent>
      </Select>
    </div>
  )
}


