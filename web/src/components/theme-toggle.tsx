"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export function ThemeToggle() {
  const { setTheme } = useTheme();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="group/button inline-flex size-8 shrink-0 items-center justify-center rounded-lg outline-none hover:bg-muted hover:text-foreground dark:hover:bg-muted/50">
        <Sun className="h-[1.2rem] w-[1.2rem] scale-100 dark:scale-0" />
        <Moon className="absolute h-[1.2rem] w-[1.2rem] scale-0 dark:scale-100" />
        <span className="sr-only">Alterar tema</span>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => setTheme("light")}>
          Claro
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setTheme("dark")}>
          Escuro
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
