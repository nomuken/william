"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const navItems = [
  { href: "/", label: "Status" },
  { href: "/interfaces", label: "Interfaces" },
  { href: "/peers", label: "Peers" },
  { href: "/allowed-emails", label: "Allowed Emails" },
  { href: "/config", label: "Config" },
  { href: "/firewall", label: "Firewall" },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="flex w-64 flex-col gap-6 border-r border-neutral-800 bg-neutral-950 px-6 py-8">
      <div className="text-lg font-semibold text-neutral-100">William Admin</div>
      <nav className="flex flex-col gap-1">
        {navItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`rounded-md px-3 py-2 text-sm transition ${
                isActive
                  ? "bg-neutral-200 text-neutral-900"
                  : "text-neutral-300 hover:bg-neutral-900 hover:text-neutral-100"
              }`}
            >
              {item.label}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
