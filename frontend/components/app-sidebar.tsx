"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { ArchiveX, Command, Home, Plus, Send, Star, Trash2 } from "lucide-react"
import { NavUser } from "@/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"
import { Select, SelectContent, SelectItem, SelectSeparator, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { createClient } from "@/lib/supabase/client"
import { useWatchlists, useWatchlistById, useCreateWatchlist } from "@/hooks/use-watchlist"
import { useScreenerResults } from "@/hooks/use-historical"
import { AddStocksPopover } from "@/components/watchlist/AddStocksPopover"
import { useCompanyInfo, useSearchCompanyInfo } from "@/hooks/use-company-info"
import type { CompanyItem } from "@/components/watchlist/AddStocksPopover"
import { useAddWatchlistItem } from "@/hooks/use-watchlist"
import { formatPrice } from "@/components/watchlist/utils"

// This is sample data
const data = {
  user: {
    name: "shadcn",
    email: "m@example.com",
    avatar: "/avatars/shadcn.jpg",
  },
  navMain: [
    {
      title: "Home",
      url: "/app",
      icon: Home,
      isActive: false,
    },
    {
      title: "Watchlist",
      url: "/app/watchlist",
      icon: Star,
      isActive: true,
    },
    {
      title: "Sent",
      url: "/app/charting",
      icon: Send,
      isActive: false,
    },
    {
      title: "Junk",
      url: "#",
      icon: ArchiveX,
      isActive: false,
    },
    {
      title: "Trash",
      url: "#",
      icon: Trash2,
      isActive: false,
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const [activeItem, setActiveItem] = React.useState(data.navMain[0])
  const { setOpen } = useSidebar()
  const pathname = usePathname()
  const router = useRouter()
  
  // Sync activeItem with pathname
  React.useEffect(() => {
    const currentNavItem = data.navMain.find(item => item.url === pathname)
    if (currentNavItem) {
      setActiveItem(currentNavItem)
    } else if (pathname === "/app/charting") {
      // Handle charting page - map to "Sent" item
      const sentItem = data.navMain.find(item => item.title === "Sent")
      if (sentItem) {
        setActiveItem(sentItem)
      }
    }
  }, [pathname])
  
  // Watchlist state
  const [user, setUser] = React.useState<{ id: string } | null>(null)
  const [selectedWatchlistId, setSelectedWatchlistId] = React.useState<string>("")
  const [openAddStocksPopover, setOpenAddStocksPopover] = React.useState(false)
  const [stockSearchTerm, setStockSearchTerm] = React.useState("")
  const [openCreateWatchlistPopover, setOpenCreateWatchlistPopover] = React.useState(false)
  const [watchlistName, setWatchlistName] = React.useState("")
  
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
  const addWatchlistItem = useAddWatchlistItem()
  const createWatchlist = useCreateWatchlist()

  // Fetch screener results counts
  const { data: insideDayData } = useScreenerResults("inside_day", "all", !!user)
  const insideDayCount = insideDayData?.count ?? 0

  const { data: highestVolumeQuarterData } = useScreenerResults("high_volume_quarter", "all", !!user)
  const highestVolumeQuarterCount = highestVolumeQuarterData?.count ?? 0

  const { data: highestVolumeYearData } = useScreenerResults("high_volume_year", "all", !!user)
  const highestVolumeYearCount = highestVolumeYearData?.count ?? 0

  const { data: highestVolumeEverData } = useScreenerResults("high_volume_ever", "all", !!user)
  const highestVolumeEverCount = highestVolumeEverData?.count ?? 0

  // Fetch company info for Add Symbol popover
  const { data: allCompanyInfo = [], isLoading: loadingAllCompanies } = useCompanyInfo(
    !!user && stockSearchTerm.length === 0 && openAddStocksPopover
  )
  const { data: searchResults = [], isLoading: searchingStocks } = useSearchCompanyInfo(
    stockSearchTerm,
    stockSearchTerm.length > 0 && openAddStocksPopover
  )

  const displayStocks: CompanyItem[] = (stockSearchTerm.length > 0 ? searchResults : allCompanyInfo).map((item) => ({
    symbol: item.symbol,
    name: item.name,
    price: item.price,
    logo: item.logo,
  }))
  const isLoadingStocks = stockSearchTerm.length > 0 ? searchingStocks : loadingAllCompanies

  // Filter out built-in watchlists from user-created ones
  const builtInWatchlistNames = ["inside day", "highest volume quarter", "highest volume year", "highest volume ever"]
  const userWatchlists = watchlists.filter(
    (w) => !builtInWatchlistNames.includes(w.name.toLowerCase())
  )

  // Built-in watchlists with their counts - memoized to prevent unnecessary re-renders
  const builtInWatchlists = React.useMemo(() => [
    { id: "inside-day", name: "Inside days", count: insideDayCount },
    { id: "highest-volume-quarter", name: "HV1", count: highestVolumeQuarterCount },
    { id: "highest-volume-year", name: "HVIPO", count: highestVolumeYearCount },
    { id: "highest-volume-ever", name: "HVE", count: highestVolumeEverCount },
  ], [insideDayCount, highestVolumeQuarterCount, highestVolumeYearCount, highestVolumeEverCount])
  
  // Fetch selected watchlist details
  const { data: selectedWatchlist, isLoading: selectedWatchlistLoading } = useWatchlistById(
    selectedWatchlistId,
    !!selectedWatchlistId && !builtInWatchlists.some(w => w.id === selectedWatchlistId)
  )

  // Handle adding a stock to the selected watchlist
  const handleAddStock = async (company: CompanyItem) => {
    if (!selectedWatchlistId || !user) return

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
    } catch (error) {
      console.error("Failed to add stock to watchlist:", error)
    }
  }

  // Handle creating a new watchlist
  const handleCreateWatchlist = async () => {
    if (!watchlistName.trim() || !user) return

    try {
      const newWatchlist = await createWatchlist.mutateAsync(watchlistName.trim())
      setWatchlistName("")
      setOpenCreateWatchlistPopover(false)
      // Select the newly created watchlist
      if (newWatchlist) {
        setSelectedWatchlistId(newWatchlist.id)
      }
    } catch (error) {
      console.error("Failed to create watchlist:", error)
    }
  }

  // Determine selected watchlist from URL query params or set default
  React.useEffect(() => {
    if ((pathname === "/app/watchlist" || pathname === "/app" || pathname === "/app/charting") && typeof window !== "undefined") {
      const params = new URLSearchParams(window.location.search)
      const watchlistParam = params.get("watchlist")
      if (watchlistParam) {
        setSelectedWatchlistId(watchlistParam)
      } else if (watchlists.length > 0 && !selectedWatchlistId && builtInWatchlists.length > 0) {
        // Set first built-in watchlist as default if available
        setSelectedWatchlistId(builtInWatchlists[0].id)
      } else if (watchlists.length > 0 && !selectedWatchlistId) {
        // Set first user watchlist as selected if on home/watchlist/charting page and no param
        setSelectedWatchlistId(watchlists[0].id)
      }
    }
  }, [pathname, watchlists, selectedWatchlistId, builtInWatchlists])
  
  // Get all watchlist options for the select
  const allWatchlistOptions = [
    ...builtInWatchlists.map(w => ({ id: w.id, name: w.name, count: w.count, isBuiltIn: true })),
    ...userWatchlists.map(w => ({ id: w.id, name: w.name, count: w.items?.length || 0, isBuiltIn: false }))
  ]
  
  // Get items to display based on selected watchlist
  const displayItems = React.useMemo(() => {
    if (!selectedWatchlistId) return []
    
    // Check if it's a built-in watchlist
    const isBuiltIn = builtInWatchlists.some(w => w.id === selectedWatchlistId)
    if (isBuiltIn) {
      // For built-in watchlists, we'd need to fetch from screener - for now return empty
      // This would need integration with the screener components
      return []
    }
    
    // For user-created watchlists, return items from selectedWatchlist
    return selectedWatchlist?.items || []
  }, [selectedWatchlistId, selectedWatchlist, builtInWatchlists])

  return (
    <Sidebar
      collapsible="icon"
      className="overflow-hidden [&>[data-sidebar=sidebar]]:flex-row"
      {...props}
    >
      {/* First sidebar - Icon navigation */}
      <Sidebar
        collapsible="none"
        className="!w-[calc(var(--sidebar-width-icon)_+_1px)] border-r"
      >
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild className="md:h-8 md:p-0">
                <a href="#">
                  <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                    <Command className="size-4" />
                  </div>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-semibold">Acme Inc</span>
                    <span className="truncate text-xs">Enterprise</span>
                  </div>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent className="px-1.5 md:px-0">
              <SidebarMenu>
                {data.navMain.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton
                      tooltip={{
                        children: item.title,
                        hidden: false,
                      }}
                      asChild={item.url !== "#"}
                      onClick={() => {
                        if (item.url === "#") {
                        setActiveItem(item)
                        setOpen(true)
                        }
                      }}
                      isActive={activeItem.title === item.title}
                      className="px-2.5 md:px-2"
                    >
                      {item.url !== "#" ? (
                        <Link href={item.url}>
                          <item.icon />
                          <span>{item.title}</span>
                        </Link>
                      ) : (
                        <>
                      <item.icon />
                      <span>{item.title}</span>
                        </>
                      )}
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <NavUser user={data.user} />
        </SidebarFooter>
      </Sidebar>

      {/* Second sidebar - Watchlist */}
      {(activeItem.title === "Watchlist" || activeItem.title === "Home" || activeItem.title === "Sent") && user ? (
        <Sidebar collapsible="none" className="hidden flex-1 md:flex">
          <SidebarHeader className="gap-3.5 border-b p-4">
            <div className="flex w-full items-center justify-between">
              <div className="text-base font-medium text-foreground">
                Watchlists
              </div>
            </div>
            <Select
              value={selectedWatchlistId || undefined}
              onValueChange={setSelectedWatchlistId}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Select a watchlist">
                  {watchlistsLoading
                    ? "Loading..."
                    : selectedWatchlistId
                    ? (() => {
                        const selected = allWatchlistOptions.find(w => w.id === selectedWatchlistId)
                        return selected ? `${selected.name} (${selected.count})` : "Select a watchlist"
                      })()
                    : "Select a watchlist"}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                {builtInWatchlists.map((watchlist) => (
                  <SelectItem key={watchlist.id} value={watchlist.id}>
                    {watchlist.name} ({watchlist.count})
                  </SelectItem>
                ))}
                {userWatchlists.map((watchlist) => {
                  const itemCount = watchlist.items?.length || 0
                  return (
                    <SelectItem key={watchlist.id} value={watchlist.id}>
                      {watchlist.name} ({itemCount})
                    </SelectItem>
                  )
                })}
                {!watchlistsLoading && allWatchlistOptions.length === 0 && (
                  <div className="px-2 py-1.5 text-sm text-muted-foreground">
                    No watchlists yet
                  </div>
                )}
                <SelectSeparator />
                <Popover open={openCreateWatchlistPopover} onOpenChange={setOpenCreateWatchlistPopover}>
                  <PopoverTrigger asChild>
                    <div className="px-2 py-1.5">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start gap-2 h-auto py-1.5"
                        onClick={(e) => {
                          e.preventDefault()
                          e.stopPropagation()
                          setOpenCreateWatchlistPopover(true)
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
                            setOpenCreateWatchlistPopover(false)
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
              </SelectContent>
            </Select>
          </SidebarHeader>
          <SidebarContent className="flex flex-col">
            {selectedWatchlistId && (
              <div className="p-4">
                <AddStocksPopover
                  open={openAddStocksPopover}
                  onOpenChange={setOpenAddStocksPopover}
                  stockSearchTerm={stockSearchTerm}
                  setStockSearchTerm={setStockSearchTerm}
                  isLoadingStocks={isLoadingStocks}
                  displayStocks={displayStocks}
                  onAddStock={handleAddStock}
                  isAdding={addWatchlistItem.isPending}
                  triggerLabel="Add Symbol"
                  widthClass="w-80"
                />
              </div>
            )}
            <div className="flex-1 overflow-y-auto">
              {selectedWatchlistLoading ? (
                <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                  Loading...
                </div>
              ) : displayItems.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8 px-4">
                  <p className="text-sm text-muted-foreground text-center">
                    {selectedWatchlistId ? "No symbols in this watchlist" : "Select a watchlist to view symbols"}
                  </p>
                </div>
              ) : (
                <div className="divide-y">
                  {displayItems.map((item) => {
                    const change = item.change || 0
                    const percentChange = item.percentChange || "0"
                    const changeNum = parseFloat(percentChange)
                    const isPositive = changeNum >= 0
                    const changeColor = isPositive ? "text-green-500" : "text-orange-500"
                    const changeSign = isPositive ? "+" : ""
                    
                    return (
                      <div
                        key={item.id}
                        className="px-4 py-3 hover:bg-sidebar-accent/50 transition-colors cursor-pointer border-b last:border-b-0"
                        onClick={() => {
                          const symbol = item.symbol || item.name.split(" ")[0]
                          // Navigate to charting if on charting page, otherwise to watchlist
                          if (pathname === "/app/charting") {
                            router.push(`/app/charting?symbol=${encodeURIComponent(symbol)}&watchlist=${selectedWatchlistId}`)
                          } else {
                            router.push(`/app/watchlist?symbol=${encodeURIComponent(symbol)}&watchlist=${selectedWatchlistId}`)
                          }
                        }}
                      >
                        <div className="flex items-start justify-between gap-2">
                          {/* Left side - Symbol and Name */}
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                              <span className="font-semibold text-sm text-foreground whitespace-nowrap">
                                {item.symbol || item.name.split(" ")[0]}
                              </span>
                              {item.starred && (
                                <span className="text-xs">⭐</span>
                              )}
                            </div>
                            <div className="text-xs text-muted-foreground truncate">
                              {item.name}
                            </div>
                          </div>

                          {/* Right side - Price and Change */}
                          <div className="flex flex-col items-end gap-1 flex-shrink-0">
                            <span className="font-semibold text-sm text-foreground">
                              ${formatPrice(item.price)}
                            </span>
                            <span className={`text-xs ${changeColor}`}>
                              {changeSign}{Math.abs(change).toFixed(2)} {changeSign}{Math.abs(changeNum).toFixed(2)}%
                            </span>
                          </div>
                        </div>
                      </div>
                    ) as React.ReactElement
                  })}
                </div>
              )}
            </div>
          </SidebarContent>
        </Sidebar>
      ) : null}
    </Sidebar>
  )
}