"use client"

import * as React from "react"
import { Bell, ChevronRight, Copy, Search, Star } from "lucide-react"

export interface StockActionMenuItem {
  id: string
  symbol: string
  name: string
}

export function StockActionMenu({
  open,
  position,
  selectedItem,
  onClose,
  onDelete,
  onMoveTop,
  pendingDelete,
}: {
  open: boolean
  position: { x: number; y: number } | null
  selectedItem: StockActionMenuItem | null
  onClose: () => void
  onDelete: () => void
  onMoveTop: () => void
  pendingDelete?: boolean
}) {
  if (!open || !position || !selectedItem) return null

  return (
    <div
      className="fixed inset-0 z-50"
      onClick={onClose}
    >
      <div
        className="absolute bg-popover text-popover-foreground rounded-md border shadow-lg w-64"
        style={{
          right: `calc(100vw - ${position.x}px + 16px)`,
          top: `${position.y}px`,
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-3">
          <h4 className="font-medium text-sm mb-3">Color Marking</h4>
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
            onClick={onMoveTop}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors"
          >
            Top
          </button>
          <button
            onClick={onDelete}
            disabled={!!pendingDelete}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors text-destructive disabled:opacity-50"
          >
            Delete
          </button>
          <button
            onClick={onClose}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors"
          >
            Add to Voice Quote
          </button>
          <button
            onClick={onClose}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors"
          >
            Create Order
          </button>
          <button
            onClick={onClose}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center gap-2"
          >
            <Bell className="size-4" />
            Create Alert
          </button>
          <button
            onClick={onClose}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center justify-between"
          >
            <span className="flex items-center gap-2">
              <Star className="size-4" />
              Add to Watchlist
            </span>
            <ChevronRight className="size-4" />
          </button>
          <button
            onClick={onClose}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center justify-between"
          >
            <span className="flex items-center gap-2">
              <Search className="size-4" />
              View {selectedItem.symbol} in a widget
            </span>
            <ChevronRight className="size-4" />
          </button>
          <button
            onClick={onClose}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center gap-2"
          >
            <Copy className="size-4" />
            Copy {selectedItem.symbol}
          </button>
          <button
            onClick={onClose}
            className="w-full px-3 py-2 text-sm text-left hover:bg-accent transition-colors flex items-center justify-between"
          >
            <span>Send {selectedItem.symbol} to</span>
            <ChevronRight className="size-4" />
          </button>
        </div>
      </div>
    </div>
  )
}


