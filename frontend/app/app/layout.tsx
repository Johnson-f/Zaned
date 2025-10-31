import { SidebarLeft } from "@/components/sidebar-left"
import { SidebarRight } from "@/components/sidebar-right"
import { DynamicHeader } from "@/components/dynamic-header"
import {
  SidebarInset,
  SidebarProvider,
} from "@/components/ui/sidebar"

export default function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <SidebarProvider className="flex flex-row">
      <SidebarLeft />
      <SidebarInset className="flex flex-col min-w-0 flex-1">
        <DynamicHeader />
        <div className="flex flex-1 flex-col gap-4 p-4 overflow-auto">
          {children}
        </div>
      </SidebarInset>
      <SidebarRight />
    </SidebarProvider>
  );
}
