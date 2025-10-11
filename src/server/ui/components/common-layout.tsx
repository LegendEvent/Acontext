"use client";

import { AppSidebar } from "@/components/app-sidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { ThemeToggle } from "@/components/theme-toggle";
import { LangSwitch } from "@/components/lang-switch";

export default function CommonLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <SidebarProvider defaultOpen={false} open={false}>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-12 shrink-0 items-center gap-2 border-b">
          <div className="flex-1 flex items-center gap-2 px-3">
            <div className="flex-1" />
            <ThemeToggle />
            <LangSwitch />
          </div>
        </header>
        {children}
      </SidebarInset>
    </SidebarProvider>
  );
}
