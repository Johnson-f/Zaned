import * as React from "react"

import { NavUser } from "@/components/nav-user"
import { Watchlist } from "@/components/watchlist"
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"

export function SidebarRight({
  ...props
}: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar
      collapsible="none"
      className="sticky top-0 hidden h-svh border-l lg:flex w-70"
      {...props}
    >
      <SidebarHeader className="border-sidebar-border h-16 border-b">
        <NavUser />
      </SidebarHeader>
      <SidebarContent className="p-0">
        <Watchlist />
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}
