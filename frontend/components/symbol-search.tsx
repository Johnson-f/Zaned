"use client"

import * as React from "react"
import { Search, X } from "lucide-react"
import { useRouter } from "next/navigation"
import { useSearchCompanyInfo } from "@/hooks/use-company-info"
import { Input } from "@/components/ui/input"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/utils"

export function SymbolSearch() {
  const [searchTerm, setSearchTerm] = React.useState("")
  const [open, setOpen] = React.useState(false)
  const [selectedIndex, setSelectedIndex] = React.useState(0)
  const inputRef = React.useRef<HTMLInputElement>(null)
  const router = useRouter()

  // Search for stocks
  const { data: searchResults = [], isLoading } = useSearchCompanyInfo(
    searchTerm,
    searchTerm.length > 0 && open
  )

  // Filter results to show up to 10
  const displayResults = searchResults.slice(0, 10)

  // Handle keyboard navigation
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault()
      setSelectedIndex((prev) => 
        prev < displayResults.length - 1 ? prev + 1 : prev
      )
    } else if (e.key === "ArrowUp") {
      e.preventDefault()
      setSelectedIndex((prev) => (prev > 0 ? prev - 1 : 0))
    } else if (e.key === "Enter") {
      e.preventDefault()
      if (displayResults[selectedIndex]) {
        handleSelectStock(displayResults[selectedIndex].symbol)
      }
    } else if (e.key === "Escape") {
      setOpen(false)
      setSearchTerm("")
    }
  }

  // Handle selecting a stock
  const handleSelectStock = (symbol: string) => {
    // Navigate to screener page with symbol query param or handle selection
    router.push(`/app/screener?symbol=${symbol}`)
    setOpen(false)
    setSearchTerm("")
    setSelectedIndex(0)
  }

  // Reset selected index when search term changes
  React.useEffect(() => {
    setSelectedIndex(0)
  }, [searchTerm])

  // Focus input when popover opens
  React.useEffect(() => {
    if (open && inputRef.current) {
      inputRef.current.focus()
    }
  }, [open])

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className="flex items-center gap-2 px-3 py-2 rounded-md border bg-background hover:bg-accent transition-colors text-sm text-muted-foreground hover:text-foreground"
          onClick={() => setOpen(true)}
        >
          <Search className="size-4" />
          <span className="hidden sm:inline">Search stocks...</span>
          <span className="sm:hidden">Search...</span>
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-96 p-0" align="start">
        <div className="flex items-center border-b px-3">
          <Search className="size-4 text-muted-foreground mr-2" />
          <Input
            ref={inputRef}
            placeholder="Search by symbol or name..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyDown={handleKeyDown}
            className="border-0 focus-visible:ring-0 focus-visible:ring-offset-0 h-12"
            autoFocus
          />
          {searchTerm && (
            <button
              onClick={() => {
                setSearchTerm("")
                inputRef.current?.focus()
              }}
              className="ml-2 p-1 hover:bg-accent rounded-md"
            >
              <X className="size-4" />
            </button>
          )}
        </div>
        {searchTerm.length > 0 && (
          <div className="max-h-96 overflow-y-auto">
            {isLoading ? (
              <div className="px-4 py-8 text-sm text-center text-muted-foreground">
                Searching...
              </div>
            ) : displayResults.length === 0 ? (
              <div className="px-4 py-8 text-sm text-center text-muted-foreground">
                No stocks found
              </div>
            ) : (
              <div className="py-1">
                {displayResults.map((stock, index) => {
                  const price = stock.price ? parseFloat(stock.price.replace(/[^0-9.-]/g, '')) : null
                  
                  return (
                    <button
                      key={stock.symbol}
                      onClick={() => handleSelectStock(stock.symbol)}
                      className={cn(
                        "w-full px-4 py-3 text-left hover:bg-accent transition-colors flex items-center gap-3",
                        index === selectedIndex && "bg-accent"
                      )}
                    >
                      {/* Logo */}
                      {stock.logo && (
                        <img
                          src={stock.logo}
                          alt={stock.symbol}
                          className="size-8 rounded object-contain"
                          onError={(e) => {
                            e.currentTarget.style.display = 'none'
                          }}
                        />
                      )}
                      {/* Stock Info */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-sm">{stock.symbol}</span>
                          {price !== null && (
                            <span className="text-xs text-muted-foreground">
                              ${price.toFixed(2)}
                            </span>
                          )}
                        </div>
                        <div className="text-xs text-muted-foreground truncate">
                          {stock.name}
                        </div>
                      </div>
                    </button>
                  )
                })}
              </div>
            )}
          </div>
        )}
        {searchTerm.length === 0 && (
          <div className="px-4 py-8 text-sm text-center text-muted-foreground">
            Start typing to search for stocks...
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}

