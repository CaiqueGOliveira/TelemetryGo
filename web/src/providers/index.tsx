import { ThemeProvider } from "@/providers/theme-provider";
import { AxiosProvider } from "@/providers/axios-provider";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="dark"
      enableSystem={false}
      disableTransitionOnChange
    >
      <AxiosProvider>{children}</AxiosProvider>
    </ThemeProvider>
  );
}
