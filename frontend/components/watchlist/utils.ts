export function formatAfterHours(percentChange?: string | null) {
  if (!percentChange) return null
  const num = parseFloat(percentChange)
  const sign = num >= 0 ? "+" : ""
  return `${sign}${num.toFixed(2)}%`
}

export function getAfterHoursColor(percentChange?: string | null) {
  if (!percentChange) return "text-muted-foreground"
  const num = parseFloat(percentChange)
  return num >= 0 ? "text-green-500" : "text-orange-500"
}

export function formatPrice(price?: number | null) {
  if (price === null || price === undefined) return "N/A"
  return price.toFixed(2)
}


