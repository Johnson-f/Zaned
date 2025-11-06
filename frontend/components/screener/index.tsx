"use client"

import * as React from "react"
import { useSearchParams } from "next/navigation"
import { Charting } from "@/components/screener/tab/charts/charting"
import { SymbolSearch } from "@/components/symbol-search"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export function Screener() {
  const searchParams = useSearchParams()
  const symbol = searchParams.get("symbol") || ""
  const [range, setRange] = React.useState("1d")
  const [interval, setInterval] = React.useState("1m")

  // Default ranges and intervals mapping
  const rangeIntervalMap: Record<string, { range: string; interval: string }> = {
    "1d": { range: "1d", interval: "1m" },
    "5d": { range: "5d", interval: "5m" },
    "1mo": { range: "1mo", interval: "15m" },
    "3mo": { range: "3mo", interval: "30m" },
    "6mo": { range: "6mo", interval: "1h" },
    "ytd": { range: "ytd", interval: "1d" },
    "1y": { range: "1y", interval: "1mo" },
    "2y": { range: "2y", interval: "1mo" },
    "5y": { range: "5y", interval: "1mo" },
    "10y": { range: "10y", interval: "1mo" },
    "max": { range: "max", interval: "1mo" },
  }

  const handleRangeChange = (value: string) => {
    const mapping = rangeIntervalMap[value]
    if (mapping) {
      setRange(mapping.range)
      setInterval(mapping.interval)
    }
  }

  if (!symbol) {
    return (
      <div className="flex-1 w-full flex flex-col gap-6">
        <div className="flex flex-col gap-2 items-start">
          <h2 className="font-bold text-2xl mb-4">Screener</h2>
          <p className="text-muted-foreground mb-6">
            Search for a stock symbol to view its chart and historical data.
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Search for a Stock</CardTitle>
            <CardDescription>
              Use the search bar in the header or search below to find a stock symbol
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="max-w-md">
              <SymbolSearch />
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex-1 w-full flex flex-col gap-4">
      {/* Range Selection Tabs */}
      <Tabs value={range} onValueChange={handleRangeChange} className="w-full">
        <TabsList className="grid w-full max-w-2xl grid-cols-6 lg:grid-cols-11">
          <TabsTrigger value="1d">1D</TabsTrigger>
          <TabsTrigger value="5d">5D</TabsTrigger>
          <TabsTrigger value="1mo">1M</TabsTrigger>
          <TabsTrigger value="3mo">3M</TabsTrigger>
          <TabsTrigger value="6mo">6M</TabsTrigger>
          <TabsTrigger value="ytd">YTD</TabsTrigger>
          <TabsTrigger value="1y">1Y</TabsTrigger>
          <TabsTrigger value="2y">2Y</TabsTrigger>
          <TabsTrigger value="5y">5Y</TabsTrigger>
          <TabsTrigger value="10y">10Y</TabsTrigger>
          <TabsTrigger value="max">MAX</TabsTrigger>
        </TabsList>

        {/* Render Charting component for each tab value - only active one is rendered */}
        {(Object.keys(rangeIntervalMap) as Array<keyof typeof rangeIntervalMap>).map((tabValue) => {
          const mapping = rangeIntervalMap[tabValue]
          return (
            <TabsContent key={tabValue} value={tabValue} className="mt-4">
              <Charting symbol={symbol} range={mapping.range} interval={mapping.interval} />
            </TabsContent>
          )
        })}
      </Tabs>
    </div>
  )
}

