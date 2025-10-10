import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/theme-toggle";
import Image from "next/image";

export default function Home() {
  return (
    <div className="flex flex-col items-center justify-center h-screen space-y-4">
      <Image
        className="dark:invert"
        src="/logo_black.svg"
        alt="Acontext logo"
        width={180}
        height={180}
        priority
      />
      <Button>Hello World</Button>
      <ThemeToggle />
    </div>
  );
}
