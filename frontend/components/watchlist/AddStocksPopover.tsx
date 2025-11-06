"use client"

import * as React from "react"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Plus, Star } from "lucide-react"

export interface CompanyItem {
  symbol: string
  name: string
  price?: string
  logo?: string
}

export function AddStocksPopover({
  open,
  onOpenChange,
  stockSearchTerm,
  setStockSearchTerm,
  isLoadingStocks,
  displayStocks,
  onAddStock,
  isAdding,
  triggerLabel = "Add stocks",
  widthClass = "w-96",
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  stockSearchTerm: string
  setStockSearchTerm: (val: string) => void
  isLoadingStocks: boolean
  displayStocks: CompanyItem[]
  onAddStock: (company: CompanyItem) => void
  isAdding: boolean
  triggerLabel?: string
  widthClass?: string
}) {
  return (
    <Popover open={open} onOpenChange={onOpenChange}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className="gap-2"
          onClick={() => onOpenChange(true)}
        >
          <Plus className="size-4" />
          <span>{triggerLabel}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className={widthClass}>
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
                  return (
                    <div
                      key={company.symbol}
                      className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-accent transition-colors"
                    >
                      {company.logo && (
                        <img
                          src={company.logo}
                          alt={company.symbol}
                          className="size-8 rounded object-contain"
                          onError={(e) => { (e.currentTarget as HTMLImageElement).style.display = 'none' }}
                        />
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="font-medium text-sm">{company.symbol}</div>
                        <div className="text-xs text-muted-foreground truncate">
                          {company.name}
                        </div>
                      </div>
                      <div className="text-sm font-medium">
                        {price !== null ? `$${price.toFixed(2)}` : 'N/A'}
                      </div>
                      <button
                        onClick={() => onAddStock(company)}
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
                onOpenChange(false)
                setStockSearchTerm("")
              }}
            >
              Close
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}


