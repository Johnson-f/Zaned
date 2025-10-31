"use client"

import { usePathname } from "next/navigation"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"

// Map paths to page titles
const pathToTitleMap: Record<string, string> = {
  "/app": "Home",
  "/app/ai": "Ask AI",
  "/app/screener": "Screener",
  "/app/watchlist": "Watchlist",
}

// Fallback function to generate title from path
function getPageTitle(pathname: string): string {
  // Check exact match first
  if (pathToTitleMap[pathname]) {
    return pathToTitleMap[pathname]
  }

  // Fallback: capitalize and format the path segment
  const segments = pathname.split("/").filter(Boolean)
  const lastSegment = segments[segments.length - 1]
  
  if (!lastSegment) {
    return "Home"
  }

  // Capitalize first letter and replace hyphens/underscores with spaces
  return lastSegment
    .replace(/[-_]/g, " ")
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

export function DynamicHeader() {
  const pathname = usePathname()
  const pageTitle = getPageTitle(pathname)

  return (
    <header className="bg-background sticky top-0 z-10 flex h-16 shrink-0 items-center gap-2 border-b">
      <div className="flex flex-1 items-center gap-2 px-3">
        <SidebarTrigger />
        <Separator
          orientation="vertical"
          className="mr-2 data-[orientation=vertical]:h-4"
        />
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbPage className="line-clamp-1 text-2xl font-bold">
                {pageTitle}
              </BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
      </div>
    </header>
  )
}

