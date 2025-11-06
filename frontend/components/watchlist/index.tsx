"use client"

import * as React from "react"
import { List, Plus, Star } from "lucide-react"
import { createClient } from "@/lib/supabase/client"
import { WatchlistSelect } from "@/components/watchlist/WatchlistSelect"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { AddStocksPopover } from "@/components/watchlist/AddStocksPopover"
import { StockActionMenu } from "@/components/watchlist/StockActionMenu"
import {
  useWatchlists,
  useWatchlistById,
  useCreateWatchlist,
  useAddWatchlistItem,
  useDeleteWatchlistItem,
} from "@/hooks/use-watchlist"
import { useCompanyInfo, useSearchCompanyInfo } from "@/hooks/use-company-info"
import { useQueryClient } from "@tanstack/react-query"
import { watchlistKeys } from "@/hooks/use-watchlist"
import type { Watchlist } from "@/lib/types/watchlist"
import { formatAfterHours, getAfterHoursColor, formatPrice } from "@/components/watchlist/utils"
import { useInsideDaySymbols, useHighVolumeQuarterSymbols, useHighVolumeYearSymbols, useHighVolumeEverSymbols } from "@/hooks/use-historical"
import InsideDayWatchlist from "@/components/watchlist/inside-day-watchlist"
import HighestVolumeQuarterWatchlist from "@/components/watchlist/highest-volume-quarter"
import HighestVolumeYearWatchlist from "@/components/watchlist/highest-volume-year"
import HighestVolumeEverWatchlist from "@/components/watchlist/highest-volume-ever"

const INSIDE_DAY_VALUE = "inside-day"
const HIGHEST_VOLUME_QUARTER_VALUE = "highest-volume-quarter"
const HIGHEST_VOLUME_YEAR_VALUE = "highest-volume-year"
const HIGHEST_VOLUME_EVER_VALUE = "highest-volume-ever"

export function Watchlist() {
  const [user, setUser] = React.useState<{ id: string } | null>(null)
  const [selectedWatchlistId, setSelectedWatchlistId] = React.useState<string>("")
  const [openPopover, setOpenPopover] = React.useState(false)
  const [watchlistName, setWatchlistName] = React.useState("")
  const [openAddStocksPopover, setOpenAddStocksPopover] = React.useState(false)
  const [stockSearchTerm, setStockSearchTerm] = React.useState("")
  const [selectedItem, setSelectedItem] = React.useState<{ id: string; symbol: string; name: string } | null>(null)
  const [openItemMenu, setOpenItemMenu] = React.useState(false)
  const [menuPosition, setMenuPosition] = React.useState<{ x: number; y: number } | null>(null)
  const queryClient = useQueryClient()

  // Check if user is authenticated
  React.useEffect(() => {
    async function checkAuth() {
      const supabase = createClient()
      const { data: { user: authUser } } = await supabase.auth.getUser()
      
      if (authUser) {
        setUser({ id: authUser.id })
      }
    }

    checkAuth()

    // Listen for auth changes
    const supabase = createClient()
    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      (_event, session) => {
        if (session?.user) {
          setUser({ id: session.user.id })
        } else {
          setUser(null)
        }
      }
    )

    return () => {
      subscription.unsubscribe()
    }
  }, [])

  // Fetch watchlists if user is authenticated
  const { data: watchlists = [], isLoading: watchlistsLoading } = useWatchlists(!!user)
  const createWatchlist = useCreateWatchlist()
  const addWatchlistItem = useAddWatchlistItem()
  const deleteWatchlistItem = useDeleteWatchlistItem()

  // Fetch inside day symbols count
  const { data: insideDayData } = useInsideDaySymbols(!!user)
  const insideDayCount = insideDayData?.count ?? 0

  // Fetch highest volume quarter symbols count
  const { data: highestVolumeQuarterData } = useHighVolumeQuarterSymbols(!!user)
  const highestVolumeQuarterCount = highestVolumeQuarterData?.count ?? 0

  // Fetch highest volume year symbols count
  const { data: highestVolumeYearData } = useHighVolumeYearSymbols(!!user)
  const highestVolumeYearCount = highestVolumeYearData?.count ?? 0

  // Fetch highest volume ever symbols count
  const { data: highestVolumeEverData } = useHighVolumeEverSymbols(!!user)
  const highestVolumeEverCount = highestVolumeEverData?.count ?? 0

  // Ensure "Inside Day" watchlist exists (only once, at parent level)
  React.useEffect(() => {
    if (!user || watchlistsLoading) return
    
    const BUILT_IN_WATCHLIST_NAME = "Inside Day"
    const existing = watchlists.find((w) => w.name.toLowerCase() === BUILT_IN_WATCHLIST_NAME.toLowerCase())
    
    if (!existing) {
      // Only create if it doesn't exist
      createWatchlist.mutate(BUILT_IN_WATCHLIST_NAME)
    }
  }, [user, watchlists, watchlistsLoading, createWatchlist])

  // Ensure "Highest Volume Quarter" watchlist exists (only once, at parent level)
  React.useEffect(() => {
    if (!user || watchlistsLoading) return
    
    const BUILT_IN_WATCHLIST_NAME = "Highest Volume Quarter"
    const existing = watchlists.find((w) => w.name.toLowerCase() === BUILT_IN_WATCHLIST_NAME.toLowerCase())
    
    if (!existing) {
      // Only create if it doesn't exist
      createWatchlist.mutate(BUILT_IN_WATCHLIST_NAME)
    }
  }, [user, watchlists, watchlistsLoading, createWatchlist])

  // Ensure "Highest Volume Year" watchlist exists (only once, at parent level)
  React.useEffect(() => {
    if (!user || watchlistsLoading) return
    
    const BUILT_IN_WATCHLIST_NAME = "Highest Volume Year"
    const existing = watchlists.find((w) => w.name.toLowerCase() === BUILT_IN_WATCHLIST_NAME.toLowerCase())
    
    if (!existing) {
      // Only create if it doesn't exist
      createWatchlist.mutate(BUILT_IN_WATCHLIST_NAME)
    }
  }, [user, watchlists, watchlistsLoading, createWatchlist])

  // Ensure "Highest Volume Ever" watchlist exists (only once, at parent level)
  React.useEffect(() => {
    if (!user || watchlistsLoading) return
    
    const BUILT_IN_WATCHLIST_NAME = "Highest Volume Ever"
    const existing = watchlists.find((w) => w.name.toLowerCase() === BUILT_IN_WATCHLIST_NAME.toLowerCase())
    
    if (!existing) {
      // Only create if it doesn't exist
      createWatchlist.mutate(BUILT_IN_WATCHLIST_NAME)
    }
  }, [user, watchlists, watchlistsLoading, createWatchlist])

  // Fetch all company info or search results
  const { data: allCompanyInfo = [], isLoading: loadingAllCompanies } = useCompanyInfo(
    !!user && stockSearchTerm.length === 0 && openAddStocksPopover
  )
  const { data: searchResults = [], isLoading: searchingStocks } = useSearchCompanyInfo(
    stockSearchTerm,
    stockSearchTerm.length > 0 && openAddStocksPopover
  )

  // Determine which stocks to display
  const displayStocks = stockSearchTerm.length > 0 ? searchResults : allCompanyInfo
  const isLoadingStocks = stockSearchTerm.length > 0 ? searchingStocks : loadingAllCompanies

  // Set first watchlist as selected when watchlists load
  React.useEffect(() => {
    if (watchlists.length > 0 && !selectedWatchlistId) {
      setSelectedWatchlistId(watchlists[0].id)
    }
  }, [watchlists, selectedWatchlistId])

  // Fetch selected watchlist details
  const { data: selectedWatchlist, isLoading: watchlistLoading } = useWatchlistById(
    selectedWatchlistId,
    !!selectedWatchlistId
  )

  // Handle creating a new watchlist
  const handleCreateWatchlist = async () => {
    if (!watchlistName.trim()) return

    try {
      const newWatchlist = await createWatchlist.mutateAsync(watchlistName.trim())
      setWatchlistName("")
      setOpenPopover(false)
      // Select the newly created watchlist
      if (newWatchlist) {
        setSelectedWatchlistId(newWatchlist.id)
      }
    } catch (error) {
      console.error("Failed to create watchlist:", error)
    }
  }

  // Handle adding a stock to the watchlist
  const handleAddStock = async (company: { symbol: string; name: string; price?: string; logo?: string }) => {
    if (!selectedWatchlistId) return

    try {
      const priceNum = company.price ? parseFloat(company.price.replace(/[^0-9.-]/g, '')) : undefined
      await addWatchlistItem.mutateAsync({
        watchlistId: selectedWatchlistId,
        item: {
          symbol: company.symbol,
          name: company.name,
          price: priceNum,
          logo: company.logo,
        },
      })
      // Don't close popover or reset search - allow adding multiple stocks
    } catch (error) {
      console.error("Failed to add stock to watchlist:", error)
    }
  }

  // Handle double-click or right-click on stock item
  const handleItemAction = (e: React.MouseEvent, item: { id: string; symbol: string; name: string }) => {
    e.preventDefault()
    e.stopPropagation()
    setSelectedItem(item)
    setMenuPosition({ x: e.clientX, y: e.clientY })
    setOpenItemMenu(true)
  }

  const handleItemDoubleClick = (e: React.MouseEvent, item: { id: string; symbol: string; name: string }) => {
    handleItemAction(e, item)
  }

  const handleItemRightClick = (e: React.MouseEvent, item: { id: string; symbol: string; name: string }) => {
    e.preventDefault()
    e.stopPropagation()
    handleItemAction(e, item)
  }

  // Handle delete stock
  const handleDeleteStock = async () => {
    if (!selectedItem) return

    try {
      await deleteWatchlistItem.mutateAsync(selectedItem.id)
      setOpenItemMenu(false)
      setSelectedItem(null)
    } catch (error) {
      console.error("Failed to delete stock:", error)
    }
  }

  // Handle move stock to top
  const handleMoveToTop = async () => {
    if (!selectedItem || !selectedWatchlistId) return

    try {
      // Find the item in the current list
      const itemToMove = items.find(item => item.id === selectedItem.id)
      if (!itemToMove) {
        setOpenItemMenu(false)
        setSelectedItem(null)
        return
      }

      // Optimistically update the cache to move item to top
      const currentWatchlist = queryClient.getQueryData<Watchlist>(watchlistKeys.detail(selectedWatchlistId))
      if (currentWatchlist && currentWatchlist.items) {
        const reorderedItems = [
          itemToMove,
          ...currentWatchlist.items.filter((item) => item.id !== selectedItem.id)
        ]
        
        // Update the cache optimistically
        queryClient.setQueryData(watchlistKeys.detail(selectedWatchlistId), {
          ...currentWatchlist,
          items: reorderedItems,
        })
      }

      // Also update the watchlists list cache
      const watchlistsData = queryClient.getQueryData<Watchlist[]>(watchlistKeys.lists())
      if (watchlistsData) {
        const updatedWatchlists = watchlistsData.map((wl) => {
          if (wl.id === selectedWatchlistId && wl.items) {
            const reorderedItems = [
              itemToMove,
              ...wl.items.filter((item) => item.id !== selectedItem.id)
            ]
            return {
              ...wl,
              items: reorderedItems,
            }
          }
          return wl
        })
        queryClient.setQueryData(watchlistKeys.lists(), updatedWatchlists)
      }

      setOpenItemMenu(false)
      setSelectedItem(null)
    } catch (error) {
      console.error("Failed to move stock to top:", error)
    }
  }

  // If not authenticated, don't render
  if (!user) {
    return null
  }

  const items = selectedWatchlist?.items || []
  const itemCount = items.length

  // utilities moved to utils.ts

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b bg-sidebar-accent/50">
        <div className="flex items-center gap-2">
          <List className="size-4" />
          <span className="font-semibold text-sm">Watchlists</span>
        </div>
      </div>

      {/* Watchlist Select Dropdown */}
      <div className="px-4 py-3 border-b">
        <WatchlistSelect
          value={selectedWatchlistId || undefined}
          onValueChange={setSelectedWatchlistId}
          watchlists={watchlists}
          loading={watchlistsLoading}
          selectedWatchlist={selectedWatchlist}
          itemCount={itemCount}
          insideDayCount={insideDayCount}
          highestVolumeQuarterCount={highestVolumeQuarterCount}
          highestVolumeYearCount={highestVolumeYearCount}
          highestVolumeEverCount={highestVolumeEverCount}
          onCreateWatchlist={async (name: string) => {
            const newWatchlist = await createWatchlist.mutateAsync(name)
            if (newWatchlist) {
              setSelectedWatchlistId(newWatchlist.id)
            }
          }}
          isCreating={createWatchlist.isPending}
        />
      </div>

      {/* Stock List */}
      <div className="flex-1 overflow-y-auto">
        {selectedWatchlistId === INSIDE_DAY_VALUE ? (
          <InsideDayWatchlist />
        ) : selectedWatchlistId === HIGHEST_VOLUME_QUARTER_VALUE ? (
          <HighestVolumeQuarterWatchlist />
        ) : selectedWatchlistId === HIGHEST_VOLUME_YEAR_VALUE ? (
          <HighestVolumeYearWatchlist />
        ) : selectedWatchlistId === HIGHEST_VOLUME_EVER_VALUE ? (
          <HighestVolumeEverWatchlist />
        ) : !selectedWatchlistId ? (
          <div className="flex flex-col items-center justify-center py-8 px-4">
            <AddStocksPopover
              open={openPopover}
              onOpenChange={setOpenPopover}
              stockSearchTerm={watchlistName}
              setStockSearchTerm={setWatchlistName}
              isLoadingStocks={false}
              displayStocks={[]}
              onAddStock={() => {}}
              isAdding={false}
              triggerLabel="Add watchlist"
              widthClass="w-80"
            />
            {/* Create Watchlist content */}
            {openPopover && (
              <div className="hidden">{/* placeholder to keep behavior consistent */}</div>
            )}
            {/* We keep the original creation content rendered below the trigger */}
            <div className="hidden">
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
                          handleCreateWatchlist()
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
                        setOpenPopover(false)
                        setWatchlistName("")
                      }}
                    >
                      Cancel
                    </Button>
                    <Button
                      size="sm"
                      onClick={handleCreateWatchlist}
                      disabled={!watchlistName.trim() || createWatchlist.isPending}
                    >
                      {createWatchlist.isPending ? "Creating..." : "Create"}
                    </Button>
                  </div>
                </div>
            </div>
          </div>
        ) : watchlistLoading ? (
          <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
            Loading...
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8 px-4">
            <AddStocksPopover
              open={openAddStocksPopover}
              onOpenChange={setOpenAddStocksPopover}
              stockSearchTerm={stockSearchTerm}
              setStockSearchTerm={setStockSearchTerm}
              isLoadingStocks={isLoadingStocks}
              displayStocks={displayStocks as any}
              onAddStock={handleAddStock as any}
              isAdding={addWatchlistItem.isPending}
              triggerLabel="Add stocks"
            />
          </div>
        ) : (
          <>
            {/* Stock List */}
            <div className="divide-y">
              {items.map((item) => {
                const afterHours = formatAfterHours(item.percentChange)
                const afterHoursColor = getAfterHoursColor(item.percentChange)

                return (
                  <div
                    key={item.id}
                    className="px-4 py-3 hover:bg-sidebar-accent/50 transition-colors cursor-pointer select-none"
                    onDoubleClick={(e) => handleItemDoubleClick(e, {
                      id: item.id,
                      symbol: item.symbol || item.name.split(" ")[0],
                      name: item.name,
                    })}
                    onContextMenu={(e) => handleItemRightClick(e, {
                      id: item.id,
                      symbol: item.symbol || item.name.split(" ")[0],
                      name: item.name,
                    })}
                  >
                    <div className="flex items-start justify-between gap-2">
                      {/* Left side - Ticker and Name */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-semibold text-sm whitespace-nowrap">
                            {item.symbol || item.name.split(" ")[0]}
                          </span>
                          {/* Icons placeholder - you can add actual icons based on item data */}
                          {item.starred && (
                            <span className="text-xs">⭐</span>
                          )}
                        </div>
                        <div className="text-xs text-muted-foreground truncate">
                          {item.name}
                        </div>
                      </div>

                      {/* Right side - Price and After-hours */}
                      <div className="flex flex-col items-end gap-1 flex-shrink-0">
                        <span className="font-semibold text-sm">
                          {formatPrice(item.price)}
                        </span>
                        {afterHours && (
                          <span className={`text-xs ${afterHoursColor}`}>
                            After: {afterHours}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>

            {/* Add stocks button when fewer than 10 stocks - positioned at bottom */}
            {items.length < 10 && (
              <div className="px-4 py-2 border-t mt-auto">
                <AddStocksPopover
                  open={openAddStocksPopover}
                  onOpenChange={setOpenAddStocksPopover}
                  stockSearchTerm={stockSearchTerm}
                  setStockSearchTerm={setStockSearchTerm}
                  isLoadingStocks={isLoadingStocks}
                  displayStocks={displayStocks as any}
                  onAddStock={handleAddStock as any}
                  isAdding={addWatchlistItem.isPending}
                />
              </div>
            )}

           {/* Stock Action Menu Modal */}
           <StockActionMenu
             open={!!(selectedItem && openItemMenu && menuPosition)}
             position={menuPosition}
             selectedItem={selectedItem}
             onClose={() => setOpenItemMenu(false)}
             onDelete={handleDeleteStock}
             onMoveTop={handleMoveToTop}
             pendingDelete={deleteWatchlistItem.isPending}
           />
          </>
        )}
      </div>
    </div>
  )
}

