"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  useHistorical,
  useHistoricalCount,
  useStocksVolumeMetrics,
  useHistoricalBySymbol,
} from "@/hooks/use-historical";
import { Loader2 } from "lucide-react";
import { useState } from "react";

export function HistoricalTest() {
  const [selectedSymbol, setSelectedSymbol] = useState<string>("TSLA");

  const {
    data: historical,
    isLoading: isLoadingHistorical,
    error: historicalError,
  } = useHistorical();

  const {
    data: count,
    isLoading: isLoadingCount,
  } = useHistoricalCount();

  const {
    data: volumeMetrics,
    isLoading: isLoadingVolumeMetrics,
  } = useStocksVolumeMetrics();

  const {
    data: historicalBySymbol,
    isLoading: isLoadingBySymbol,
  } = useHistoricalBySymbol(selectedSymbol);

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {/* Total Count Card */}
      <Card>
        <CardHeader>
          <CardTitle>Total Historical Records</CardTitle>
          <CardDescription>Number of historical records in database</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoadingCount ? (
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>Loading...</span>
            </div>
          ) : (
            <div className="text-3xl font-bold">{count?.toLocaleString()}</div>
          )}
        </CardContent>
      </Card>

      {/* Volume Metrics Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle>Volume Metrics</CardTitle>
          <CardDescription>Highest volume statistics for all stocks</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoadingVolumeMetrics ? (
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>Loading volume metrics...</span>
            </div>
          ) : volumeMetrics && volumeMetrics.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-2 font-medium">Symbol</th>
                    <th className="text-right p-2 font-medium">Highest in Year</th>
                    <th className="text-right p-2 font-medium">Highest in Quarter</th>
                    <th className="text-right p-2 font-medium">Highest Ever</th>
                  </tr>
                </thead>
                <tbody>
                  {volumeMetrics.slice(0, 10).map((metric) => (
                    <tr
                      key={metric.symbol}
                      className="border-b hover:bg-muted/50 cursor-pointer"
                      onClick={() => setSelectedSymbol(metric.symbol)}
                    >
                      <td className="p-2 font-medium">{metric.symbol}</td>
                      <td className="p-2 text-right">
                        {(metric.highest_volume_in_year / 1000000).toFixed(2)}M
                      </td>
                      <td className="p-2 text-right">
                        {(metric.highest_volume_in_quarter / 1000000).toFixed(2)}M
                      </td>
                      <td className="p-2 text-right font-semibold text-green-600">
                        {(metric.highest_volume_ever / 1000000).toFixed(2)}M
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {volumeMetrics.length > 10 && (
                <p className="text-xs text-muted-foreground mt-4 text-center">
                  Showing 10 of {volumeMetrics.length} stocks
                </p>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No volume metrics available</p>
          )}
        </CardContent>
      </Card>

      {/* Selected Symbol Historical Data Card */}
      <Card className="md:col-span-2 lg:col-span-3">
        <CardHeader>
          <CardTitle>Historical Data for {selectedSymbol}</CardTitle>
          <CardDescription>
            Showing historical price data for selected symbol
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoadingBySymbol ? (
            <div className="flex items-center justify-center gap-2 py-8">
              <Loader2 className="h-5 w-5 animate-spin" />
              <span>Loading historical data...</span>
            </div>
          ) : historicalBySymbol && historicalBySymbol.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-2 font-medium">Epoch</th>
                    <th className="text-left p-2 font-medium">Range</th>
                    <th className="text-left p-2 font-medium">Interval</th>
                    <th className="text-right p-2 font-medium">Open</th>
                    <th className="text-right p-2 font-medium">High</th>
                    <th className="text-right p-2 font-medium">Low</th>
                    <th className="text-right p-2 font-medium">Close</th>
                    <th className="text-right p-2 font-medium">Volume</th>
                  </tr>
                </thead>
                <tbody>
                  {historicalBySymbol.slice(0, 20).map((record) => (
                    <tr
                      key={record.id}
                      className="border-b hover:bg-muted/50"
                    >
                      <td className="p-2 text-muted-foreground">
                        {new Date(record.epoch * 1000).toLocaleString()}
                      </td>
                      <td className="p-2">{record.range}</td>
                      <td className="p-2">{record.interval}</td>
                      <td className="p-2 text-right">
                        ${record.open.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-green-600">
                        ${record.high.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-red-600">
                        ${record.low.toFixed(2)}
                      </td>
                      <td className="p-2 text-right font-semibold">
                        ${record.close.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-muted-foreground">
                        {(record.volume / 1000000).toFixed(2)}M
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {historicalBySymbol.length > 20 && (
                <p className="text-xs text-muted-foreground mt-4 text-center">
                  Showing 20 of {historicalBySymbol.length} records
                </p>
              )}
            </div>
          ) : (
            <div className="py-8 text-center">
              <p className="text-muted-foreground">
                No historical data found for {selectedSymbol}
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* All Historical Records Card */}
      <Card className="md:col-span-2 lg:col-span-3">
        <CardHeader>
          <CardTitle>All Historical Records</CardTitle>
          <CardDescription>
            Complete list of historical records from the database
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoadingHistorical ? (
            <div className="flex items-center justify-center gap-2 py-8">
              <Loader2 className="h-5 w-5 animate-spin" />
              <span>Loading historical records...</span>
            </div>
          ) : historicalError ? (
            <div className="py-8 text-center">
              <p className="text-destructive">
                Error: {historicalError.message}
              </p>
            </div>
          ) : historical && historical.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-2 font-medium">Symbol</th>
                    <th className="text-left p-2 font-medium">Date</th>
                    <th className="text-left p-2 font-medium">Range</th>
                    <th className="text-left p-2 font-medium">Interval</th>
                    <th className="text-right p-2 font-medium">Open</th>
                    <th className="text-right p-2 font-medium">High</th>
                    <th className="text-right p-2 font-medium">Low</th>
                    <th className="text-right p-2 font-medium">Close</th>
                    <th className="text-right p-2 font-medium">Volume</th>
                  </tr>
                </thead>
                <tbody>
                  {historical.slice(0, 20).map((record) => (
                    <tr
                      key={record.id}
                      className="border-b hover:bg-muted/50"
                    >
                      <td className="p-2 font-medium">{record.symbol}</td>
                      <td className="p-2 text-muted-foreground">
                        {new Date(record.epoch * 1000).toLocaleDateString()}
                      </td>
                      <td className="p-2">{record.range}</td>
                      <td className="p-2">{record.interval}</td>
                      <td className="p-2 text-right">
                        ${record.open.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-green-600">
                        ${record.high.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-red-600">
                        ${record.low.toFixed(2)}
                      </td>
                      <td className="p-2 text-right font-semibold">
                        ${record.close.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-muted-foreground">
                        {(record.volume / 1000000).toFixed(2)}M
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {historical.length > 20 && (
                <p className="text-xs text-muted-foreground mt-4 text-center">
                  Showing 20 of {historical.length} records
                </p>
              )}
            </div>
          ) : (
            <div className="py-8 text-center">
              <p className="text-muted-foreground">No historical records found</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

