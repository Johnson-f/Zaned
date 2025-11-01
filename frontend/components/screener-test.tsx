"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  useScreeners,
  useTopGainers,
  useMostActive,
  useScreenerCount,
} from "@/hooks/use-screener";
import { Loader2 } from "lucide-react";

export function ScreenerTest() {
  const {
    data: screeners,
    isLoading: isLoadingScreeners,
    error: screenersError,
  } = useScreeners();

  const {
    data: topGainers,
    isLoading: isLoadingTopGainers,
  } = useTopGainers(5);

  const {
    data: mostActive,
    isLoading: isLoadingMostActive,
  } = useMostActive(5);

  const {
    data: count,
    isLoading: isLoadingCount,
  } = useScreenerCount();

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {/* Total Count Card */}
      <Card>
        <CardHeader>
          <CardTitle>Total Stocks</CardTitle>
          <CardDescription>Number of stocks in database</CardDescription>
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

      {/* Top Gainers Card */}
      <Card>
        <CardHeader>
          <CardTitle>Top Gainers</CardTitle>
          <CardDescription>Highest price movement</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoadingTopGainers ? (
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>Loading...</span>
            </div>
          ) : topGainers && topGainers.length > 0 ? (
            <div className="space-y-2">
              {topGainers.slice(0, 5).map((screener) => (
                <div
                  key={screener.id}
                  className="flex items-center justify-between text-sm"
                >
                  <span className="font-medium">{screener.symbol}</span>
                  <span className="text-green-600">
                    ${screener.close.toFixed(2)}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No data available</p>
          )}
        </CardContent>
      </Card>

      {/* Most Active Card */}
      <Card>
        <CardHeader>
          <CardTitle>Most Active</CardTitle>
          <CardDescription>Highest trading volume</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoadingMostActive ? (
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>Loading...</span>
            </div>
          ) : mostActive && mostActive.length > 0 ? (
            <div className="space-y-2">
              {mostActive.slice(0, 5).map((screener) => (
                <div
                  key={screener.id}
                  className="flex items-center justify-between text-sm"
                >
                  <span className="font-medium">{screener.symbol}</span>
                  <span className="text-muted-foreground">
                    {(screener.volume / 1000000).toFixed(2)}M
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No data available</p>
          )}
        </CardContent>
      </Card>

      {/* All Screeners List Card */}
      <Card className="md:col-span-2 lg:col-span-3">
        <CardHeader>
          <CardTitle>All Stocks</CardTitle>
          <CardDescription>
            Complete list of stocks from the database
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoadingScreeners ? (
            <div className="flex items-center justify-center gap-2 py-8">
              <Loader2 className="h-5 w-5 animate-spin" />
              <span>Loading screeners...</span>
            </div>
          ) : screenersError ? (
            <div className="py-8 text-center">
              <p className="text-destructive">
                Error: {screenersError.message}
              </p>
            </div>
          ) : screeners && screeners.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-2 font-medium">Symbol</th>
                    <th className="text-right p-2 font-medium">Open</th>
                    <th className="text-right p-2 font-medium">High</th>
                    <th className="text-right p-2 font-medium">Low</th>
                    <th className="text-right p-2 font-medium">Close</th>
                    <th className="text-right p-2 font-medium">Volume</th>
                  </tr>
                </thead>
                <tbody>
                  {screeners.slice(0, 20).map((screener) => (
                    <tr key={screener.id} className="border-b hover:bg-muted/50">
                      <td className="p-2 font-medium">{screener.symbol}</td>
                      <td className="p-2 text-right">
                        ${screener.open.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-green-600">
                        ${screener.high.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-red-600">
                        ${screener.low.toFixed(2)}
                      </td>
                      <td className="p-2 text-right font-semibold">
                        ${screener.close.toFixed(2)}
                      </td>
                      <td className="p-2 text-right text-muted-foreground">
                        {(screener.volume / 1000000).toFixed(2)}M
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {screeners.length > 20 && (
                <p className="text-xs text-muted-foreground mt-4 text-center">
                  Showing 20 of {screeners.length} stocks
                </p>
              )}
            </div>
          ) : (
            <div className="py-8 text-center">
              <p className="text-muted-foreground">No screeners found</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

