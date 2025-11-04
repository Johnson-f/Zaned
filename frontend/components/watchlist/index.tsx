"use client"

import * as React from "react"
import { List, Plus, Star, Bell, Search, Copy, ChevronRight } from "lucide-react"
import { createClient } from "@/lib/supabase/client"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  useWatchlists,
  useWatchlistById,
  useCreateWatchlist,
  useAddWatchlistItem,
  useDeleteWatchlistItem,
} from "@/hooks/use-watchlist"
import { useCompanyInfo, useSearchCompanyInfo } from "@/hooks/use-company-info"

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

  // If not authenticated, don't render
  if (!user) {
    return null
  }

  const items = selectedWatchlist?.items || []
  const itemCount = items.length

  // Format after-hours percentage
  const formatAfterHours = (percentChange?: string | null) => {
    if (!percentChange) return null
    const num = parseFloat(percentChange)
    const sign = num >= 0 ? "+" : ""
    return `${sign}${num.toFixed(2)}%`
  }

  // Get color for after-hours change
  const getAfterHoursColor = (percentChange?: string | null) => {
    if (!percentChange) return "text-muted-foreground"
    const num = parseFloat(percentChange)
    return num >= 0 ? "text-green-500" : "text-orange-500"
  }

  // Format price
  const formatPrice = (price?: number | null) => {
    if (price === null || price === undefined) return "N/A"
    return price.toFixed(2)
  }

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
        <Select
          value={selectedWatchlistId || undefined}
          onValueChange={setSelectedWatchlistId}
          disabled={watchlistsLoading}
        >
          <SelectTrigger className="w-full">
            <SelectValue>
              {watchlistsLoading
                ? "Loading..."
                : watchlists.length === 0
                ? "Select watchlist"
                : selectedWatchlist
                ? `${selectedWatchlist.name} (${itemCount})`
                : "Select watchlist"}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            {watchlists.length === 0 ? (
              <div className="px-2 py-1.5 text-sm text-muted-foreground">
                No watchlists yet
              </div>
            ) : (
              watchlists.map((watchlist) => (
                <SelectItem key={watchlist.id} value={watchlist.id}>
                  {watchlist.name} ({watchlist.items?.length || 0})
                </SelectItem>
              ))
            )}
          </SelectContent>
        </Select>
      </div>

      {/* Stock List */}
      <div className="flex-1 overflow-y-auto">
        {!selectedWatchlistId ? (
          <div className="flex flex-col items-center justify-center py-8 px-4">
            <Popover open={openPopover} onOpenChange={setOpenPopover}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className="gap-2"
                  onClick={() => setOpenPopover(true)}
                >
                  <Plus className="size-4" />
                  <span>Add watchlist</span>
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-80">
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
              </PopoverContent>
            </Popover>
          </div>
        ) : watchlistLoading ? (
          <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
            Loading...
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8 px-4">
            <Popover open={openAddStocksPopover} onOpenChange={setOpenAddStocksPopover}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className="gap-2"
                  onClick={() => setOpenAddStocksPopover(true)}
                >
                  <Plus className="size-4" />
                  <span>Add stocks</span>
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-96">
                <div className="space-y-4">
                  <div className="space-y-2">
                    <h4 className="font-medium text-sm">Add Stocks</h4>
                    <p className="text-xs text-muted-foreground">
                      Search for stocks to add to your watchlist
                    </p>
                  </div>
                  <div className="space-y-2">
                    <Input
                      placeholder="Search by symbol or name..."
                      value={stockSearchTerm}
                      onChange={(e) => setStockSearchTerm(e.target.value)}
                      autoFocus
                    />
                  </div>
                  <div className="space-y-2 max-h-96 overflow-y-auto">
                    {isLoadingStocks ? (
                      <div className="text-sm text-muted-foreground text-center py-4">
                        Loading...
                      </div>
                    ) : displayStocks.length === 0 ? (
                      <div className="text-sm text-muted-foreground text-center py-4">
                        {stockSearchTerm.length > 0 ? "No stocks found" : "No stocks available"}
                      </div>
                    ) : (
                      <div className="space-y-1">
                        {displayStocks.slice(0, 100).map((company) => {
                          const price = company.price ? parseFloat(company.price.replace(/[^0-9.-]/g, '')) : null
                          const isAdding = addWatchlistItem.isPending
                          
                          return (
                            <div
                              key={company.symbol}
                              className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-accent transition-colors"
                            >
                              {/* Logo */}
                              {company.logo && (
                                <img
                                  src={company.logo}
                                  alt={company.symbol}
                                  className="size-8 rounded object-contain"
                                  onError={(e) => {
                                    e.currentTarget.style.display = 'none'
                                  }}
                                />
                              )}
                              {/* Stock Info */}
                              <div className="flex-1 min-w-0">
                                <div className="font-medium text-sm">{company.symbol}</div>
                                <div className="text-xs text-muted-foreground truncate">
                                  {company.name}
                                </div>
                              </div>
                              {/* Price */}
                              <div className="text-sm font-medium">
                                {price !== null ? `$${price.toFixed(2)}` : 'N/A'}
                              </div>
                              {/* Star Button */}
                              <button
                                onClick={() => handleAddStock(company)}
                                disabled={isAdding}
                                className="p-1.5 rounded-md hover:bg-accent transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                title="Add to watchlist"
                              >
                                <Star className="size-4 text-muted-foreground hover:text-yellow-500 transition-colors" />
                              </button>
                            </div>
                          )
                        })}
                      </div>
                    )}
                  </div>
                  <div className="flex justify-end">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setOpenAddStocksPopover(false)
                        setStockSearchTerm("")
                      }}
                    >
                      Close
                    </Button>
                  </div>
                </div>
              </PopoverContent>
            </Popover>
          </div>
        ) : (
          <>
            {/* Add stocks button when fewer than 10 stocks */}
            {items.length < 10 && (
              <div className="px-4 py-2 border-b">
                <Popover open={openAddStocksPopover} onOpenChange={setOpenAddStocksPopover}>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      size="sm"
                      className="w-full gap-2"
                      onClick={() => setOpenAddStocksPopover(true)}
                    >
                      <Plus className="size-4" />
                      <span>Add stocks</span>
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-96">
                    <div className="space-y-4">
                      <div className="space-y-2">
                        <h4 className="font-medium text-sm">Add Stocks</h4>
                        <p className="text-xs text-muted-foreground">
                          Search for stocks to add to your watchlist
                        </p>
                      </div>
                      <div className="space-y-2">
                        <Input
                          placeholder="Search by symbol or name..."
                          value={stockSearchTerm}
                          onChange={(e) => setStockSearchTerm(e.target.value)}
                          autoFocus
                        />
                      </div>
                      <div className="space-y-2 max-h-96 overflow-y-auto">
                        {isLoadingStocks ? (
                          <div className="text-sm text-muted-foreground text-center py-4">
                            Loading...
                          </div>
                        ) : displayStocks.length === 0 ? (
                          <div className="text-sm text-muted-foreground text-center py-4">
                            {stockSearchTerm.length > 0 ? "No stocks found" : "No stocks available"}
                          </div>
                        ) : (
                          <div className="space-y-1">
                            {displayStocks.slice(0, 100).map((company) => {
                              const price = company.price ? parseFloat(company.price.replace(/[^0-9.-]/g, '')) : null
                              const isAdding = addWatchlistItem.isPending
                              
                              return (
                                <div
                                  key={company.symbol}
                                  className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-accent transition-colors"
                                >
                                  {/* Logo */}
                                  {company.logo && (
                                    <img
                                      src={company.logo}
                                      alt={company.symbol}
                                      className="size-8 rounded object-contain"
                                      onError={(e) => {
                                        e.currentTarget.style.display = 'none'
                                      }}
                                    />
                                  )}
                                  {/* Stock Info */}
                                  <div className="flex-1 min-w-0">
                                    <div className="font-medium text-sm">{company.symbol}</div>
                                    <div className="text-xs text-muted-foreground truncate">
                                      {company.name}
                                    </div>
                                  </div>
                                  {/* Price */}
                                  <div className="text-sm font-medium">
                                    {price !== null ? `$${price.toFixed(2)}` : 'N/A'}
                                  </div>
                                  {/* Star Button */}
                                  <button
                                    onClick={() => handleAddStock(company)}
                                    disabled={isAdding}
                                    className="p-1.5 rounded-md hover:bg-accent transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                    title="Add to watchlist"
                                  >
                                    <Star className="size-4 text-muted-foreground hover:text-yellow-500 transition-colors" />
                                  </button>
                                </div>
                              )
                            })}
                          </div>
                        )}
                      </div>
                      <div className="flex justify-end">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setOpenAddStocksPopover(false)
                            setStockSearchTerm("")
                          }}
                        >
                          Close
                        </Button>
                      </div>
                    </div>
                  </PopoverContent>
                </Popover>
              </div>
            )}
            
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

           {/* Stock Action Menu Modal */}
           {selectedItem && openItemMenu && menuPosition && (
              <div
                className="fixed inset-0 z-50"
                onClick={() => setOpenItemMenu(false)}
              >
                <div
                  className="absolute bg-popover text-popover-foreground rounded-md border shadow-lg w-64"
                  style={{
                    right: `calc(100vw - ${menuPosition.x}px + 16px)`,
                    top: `${menuPosition.y}px`,
                  }}
                  onClick={(e) => e.stopPropagation()}
                >
                  <div className="p-3">
                    <h4 className="font-medium text-sm mb-3">Color Marking</h4>
                    {/* Color options */}
                    <div className="flex gap-2 mb-3">
                      <div className="size-4 rounded-full bg-blue-400 cursor-pointer" />
                      <div className="size-4 rounded-full bg-green-500 cursor-pointer" />
                      <div className="size-4 rounded-full bg-purple-500 cursor-pointer" />
                      <div className="size-4 rounded-full bg-orange-500 cursor-pointer" />
                      <div className="size-4 rounded-full bg-yellow-500 cursor-pointer" />
                    </div>
                  </div>
                  <div className="border-t">
                    <button
                      onClick={() => setOpenItemMenu(false)}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors"
                    >
                      Top
                    </button>
                    <button
                      onClick={handleDeleteStock}
                      disabled={deleteWatchlistItem.isPending}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors text-destructive disabled:opacity-50"
                    >
                      Delete
                    </button>
                    <button
                      onClick={() => setOpenItemMenu(false)}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors"
                    >
                      Add to Voice Quote
                    </button>
                    <button
                      onClick={() => setOpenItemMenu(false)}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors"
                    >
                      Create Order
                    </button>
                    <button
                      onClick={() => setOpenItemMenu(false)}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center gap-2"
                    >
                      <Bell className="size-4" />
                      Create Alert
                    </button>
                    <button
                      onClick={() => setOpenItemMenu(false)}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center justify-between"
                    >
                      <span className="flex items-center gap-2">
                        <Star className="size-4" />
                        Add to Watchlist
                      </span>
                      <ChevronRight className="size-4" />
                    </button>
                    <button
                      onClick={() => setOpenItemMenu(false)}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center justify-between"
                    >
                      <span className="flex items-center gap-2">
                        <Search className="size-4" />
                        View {selectedItem.symbol} in a widget
                      </span>
                      <ChevronRight className="size-4" />
                    </button>
                    <button
                      onClick={() => setOpenItemMenu(false)}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center gap-2"
                    >
                      <Copy className="size-4" />
                      Copy {selectedItem.symbol}
                    </button>
                    <button
                      onClick={() => setOpenItemMenu(false)}
                      className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center justify-between"
                    >
                      <span>Send {selectedItem.symbol} to</span>
                      <ChevronRight className="size-4" />
                    </button>
                  </div>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

