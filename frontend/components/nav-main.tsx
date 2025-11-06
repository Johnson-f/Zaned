"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { type LucideIcon } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavMain({
  items,
}: {
  items: {
    title: string
    url: string
    icon: LucideIcon
    isActive?: boolean
  }[]
}) {
  const pathname = usePathname()

  // Helper function to check if a route is active
  const isRouteActive = (url: string, currentPath: string): boolean => {
    // Exact match
    if (currentPath === url) return true
    
    // For root route (/app), only match exactly
    if (url === "/app") {
      return currentPath === "/app"
    }
    
    // For other routes, check if pathname starts with the URL
    // Note: usePathname() already strips query params, so /app/screener?symbol=AAPL becomes /app/screener
    // This handles cases like /app/screener matching /app/screener
    return currentPath.startsWith(url + "/") || currentPath === url
  }

  return (
    <SidebarMenu>
      {items.map((item) => {
        const isActive = item.isActive ?? isRouteActive(item.url, pathname)
        
        return (
          <SidebarMenuItem key={item.title}>
            <SidebarMenuButton asChild isActive={isActive}>
              <Link href={item.url}>
                <item.icon />
                <span>{item.title}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        )
      })}
    </SidebarMenu>
  )
}
