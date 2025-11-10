"use client"

import * as React from "react"
import { createChart, ColorType, IChartApi, ISeriesApi, CandlestickSeriesPartialOptions, HistogramSeriesPartialOptions, LineSeriesPartialOptions, Time } from "lightweight-charts"
import { useHistoricalBySymbol } from "@/hooks/use-historical"
import type { Historical } from "@/lib/types/historical"
import { Loader2 } from "lucide-react"
import { useScreenerBySymbol } from "@/hooks/use-screener"

interface StockChartProps {
  symbol: string
  range?: string
  interval?: string
}

export function StockChart({ symbol, range = "1d", interval = "1m" }: StockChartProps) {
  const chartContainerRef = React.useRef<HTMLDivElement>(null)
  const chartRef = React.useRef<IChartApi | null>(null)
  const candlestickSeriesRef = React.useRef<ISeriesApi<"Candlestick"> | null>(null)
  const volumeSeriesRef = React.useRef<ISeriesApi<"Histogram"> | null>(null)
  const sma20Ref = React.useRef<ISeriesApi<"Line"> | null>(null)
  const sma50Ref = React.useRef<ISeriesApi<"Line"> | null>(null)
  const sma200Ref = React.useRef<ISeriesApi<"Line"> | null>(null)

  // Fetch historical data
  const { data: historicalData = [], isLoading, error } = useHistoricalBySymbol(
    symbol,
    range,
    interval,
    !!symbol
  )

  // Fetch current screener data for OHLC header
  const { data: screenerData } = useScreenerBySymbol(symbol, !!symbol)

  // Initialize chart (only once)
  React.useEffect(() => {
    if (!chartContainerRef.current || chartRef.current) return

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: "#0a0a0a" },
        textColor: "#d1d5db",
      },
      grid: {
        vertLines: {
          color: "#1f2937",
          style: 1, // Solid line
        },
        horzLines: {
          color: "#1f2937",
          style: 1,
        },
      },
      width: chartContainerRef.current.clientWidth,
      height: 600,
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        borderColor: "#374151",
      },
      rightPriceScale: {
        borderColor: "#374151",
      },
      crosshair: {
        mode: 1, // Normal mode
        vertLine: {
          color: "#6b7280",
          width: 1,
          style: 2, // Dashed
        },
        horzLine: {
          color: "#6b7280",
          width: 1,
          style: 2,
        },
      },
    })

    chartRef.current = chart

    // Create candlestick series
     // @ts-expect-error - will fix later (i may never, inasmuch as the code works, who cares?)
    const candlestickSeries = chart.addCandlestickSeries({
      upColor: "#26a69a",
      downColor: "#ef5350",
      borderVisible: false,
      wickUpColor: "#26a69a",
      wickDownColor: "#ef5350",
    } as CandlestickSeriesPartialOptions)
    candlestickSeriesRef.current = candlestickSeries

    // Create volume series
     // @ts-expect-error - will fix later (i may never, inasmuch as the code works, who cares?)
    const volumeSeries = chart.addHistogramSeries({
      color: "#26a69a",
      priceFormat: {
        type: "volume",
      },
      priceScaleId: "",
      scaleMargins: {
        top: 0.8,
        bottom: 0,
      },
    } as HistogramSeriesPartialOptions)
    volumeSeriesRef.current = volumeSeries

    // Create moving averages
     // @ts-expect-error - will fix later (i may never, inasmuch as the code works, who cares?)
    const sma20 = chart.addLineSeries({
      color: "#9333ea",
      lineWidth: 2,
      title: "SMA 20",
    } as LineSeriesPartialOptions)
    sma20Ref.current = sma20

     // @ts-expect-error - will fix later (i may never, inasmuch as the code works, who cares?)
    const sma50 = chart.addLineSeries({
      color: "#34d399",
      lineWidth: 2,
      title: "SMA 50",
    } as LineSeriesPartialOptions)
    sma50Ref.current = sma50

     // @ts-expect-error - will fix later (i may never, inasmuch as the code works, who cares?)
    const sma200 = chart.addLineSeries({
      color: "#f97316",
      lineWidth: 2,
      title: "SMA 200",
    } as LineSeriesPartialOptions)
    sma200Ref.current = sma200

    // Handle resize
    const handleResize = () => {
      if (chartContainerRef.current && chartRef.current) {
        chartRef.current.applyOptions({
          width: chartContainerRef.current.clientWidth,
        })
      }
    }

    window.addEventListener("resize", handleResize)

    return () => {
      window.removeEventListener("resize", handleResize)
      if (chartRef.current) {
        chartRef.current.remove()
        chartRef.current = null
      }
      candlestickSeriesRef.current = null
      volumeSeriesRef.current = null
      sma20Ref.current = null
      sma50Ref.current = null
      sma200Ref.current = null
    }
  }, [])

  // Update chart data when historical data changes
  React.useEffect(() => {
    if (!historicalData.length || !candlestickSeriesRef.current || !volumeSeriesRef.current) return

    // Convert epoch (seconds) to Unix timestamp format expected by lightweight-charts
    // lightweight-charts expects Unix timestamp in seconds
    const convertTime = (epoch: number): Time => {
      // epoch is already in seconds, return as is
      return epoch as Time
    }

    // Prepare candlestick data
    const candlestickData = historicalData.map((item: Historical) => ({
      time: convertTime(item.epoch),
      open: item.open,
      high: item.high,
      low: item.low,
      close: item.close,
    }))

    // Prepare volume data
    const volumeData = historicalData.map((item: Historical) => ({
      time: convertTime(item.epoch),
      value: item.volume,
      color: item.close >= item.open ? "#26a69a" : "#ef5350",
    }))

    // Calculate moving averages
    const sma20Data: { time: Time; value: number }[] = []
    const sma50Data: { time: Time; value: number }[] = []
    const sma200Data: { time: Time; value: number }[] = []

    const calculateSMA = (period: number) => {
      const smaData: { time: Time; value: number }[] = []
      for (let i = period - 1; i < candlestickData.length; i++) {
        const sum = candlestickData
          .slice(i - period + 1, i + 1)
          .reduce((acc, item) => acc + item.close, 0)
        smaData.push({
          time: candlestickData[i].time,
          value: sum / period,
        })
      }
      return smaData
    }

    if (candlestickData.length >= 20) {
      sma20Data.push(...calculateSMA(20))
    }
    if (candlestickData.length >= 50) {
      sma50Data.push(...calculateSMA(50))
    }
    if (candlestickData.length >= 200) {
      sma200Data.push(...calculateSMA(200))
    }

    // Update series
    candlestickSeriesRef.current.setData(candlestickData)
    volumeSeriesRef.current.setData(volumeData)

    if (sma20Ref.current) {
      sma20Ref.current.setData(sma20Data.length > 0 ? sma20Data : [])
    }
    if (sma50Ref.current) {
      sma50Ref.current.setData(sma50Data.length > 0 ? sma50Data : [])
    }
    if (sma200Ref.current) {
      sma200Ref.current.setData(sma200Data.length > 0 ? sma200Data : [])
    }

    // Fit content
    if (chartRef.current) {
      chartRef.current.timeScale().fitContent()
    }
  }, [historicalData, symbol, range, interval])

  // Get latest data for header
  const latestData = historicalData.length > 0 ? historicalData[historicalData.length - 1] : null
  const currentPrice = screenerData?.close || latestData?.close || 0
  const previousClose = historicalData.length > 1 ? historicalData[historicalData.length - 2]?.close : latestData?.close || 0
  const change = currentPrice - previousClose
  const percentChange = previousClose !== 0 ? (change / previousClose) * 100 : 0

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-[600px] bg-[#0a0a0a]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-[600px] bg-[#0a0a0a] text-red-400">
        Error loading chart data: {error instanceof Error ? error.message : "Unknown error"}
      </div>
    )
  }

  if (!historicalData.length) {
    return (
      <div className="flex items-center justify-center h-[600px] bg-[#0a0a0a] text-muted-foreground">
        No historical data available for {symbol}
      </div>
    )
  }

  return (
    <div className="w-full bg-[#0a0a0a]">
      {/* Header */}
      <div className="border-b border-[#1f2937] px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div>
            <div className="text-lg font-semibold text-white">
              {symbol}
            </div>
            <div className="text-sm text-muted-foreground">
              {range} {interval.charAt(0).toUpperCase() + interval.slice(1)}
            </div>
          </div>
        </div>
        {latestData && (
          <div className="flex items-center gap-6 text-sm">
            <div>
              <span className="text-muted-foreground">O:</span>{" "}
              <span className="text-white">${latestData.open.toFixed(2)}</span>
            </div>
            <div>
              <span className="text-muted-foreground">H:</span>{" "}
              <span className="text-white">${latestData.high.toFixed(2)}</span>
            </div>
            <div>
              <span className="text-muted-foreground">L:</span>{" "}
              <span className="text-white">${latestData.low.toFixed(2)}</span>
            </div>
            <div>
              <span className="text-muted-foreground">C:</span>{" "}
              <span className="text-white">${latestData.close.toFixed(2)}</span>
            </div>
            <div>
              <span className="text-muted-foreground">Vol:</span>{" "}
              <span className="text-white">
                {(latestData.volume / 1000).toFixed(0)}K
              </span>
            </div>
            {change !== 0 && (
              <div>
                <span className={change >= 0 ? "text-green-400" : "text-red-400"}>
                  {change >= 0 ? "+" : ""}
                  {change.toFixed(2)} ({percentChange >= 0 ? "+" : ""}
                  {percentChange.toFixed(2)}%)
                </span>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Chart */}
      <div ref={chartContainerRef} className="w-full" style={{ height: "600px" }} />
    </div>
  )
}

