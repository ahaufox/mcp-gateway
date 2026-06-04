import React, { useState } from "react";
import { getAuthToken, setAuthToken } from "../utils/api";
import { useTheme } from "../context/ThemeContext";
import {
  Activity,
  ArrowRightLeft,
  History,
  LogOut,
  ShieldCheck,
  Menu,
  X,
  Sun,
  Moon
} from "lucide-react";

interface NavbarProps {
  activePage: "dashboard" | "converter" | "changelog";
}

export const Navbar: React.FC<NavbarProps> = ({ activePage }) => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const { theme, toggleTheme } = useTheme();

  const handleLogout = () => {
    setAuthToken("");
    window.location.href = "/login";
  };

  const navItems = [
    { id: "dashboard", label: "监控面板", icon: Activity, href: "/" },
    { id: "converter", label: "配置转换", icon: ArrowRightLeft, href: "/docs/" },
    { id: "changelog", label: "更新日志", icon: History, href: "/changelog/" }
  ];

  return (
    <nav className="sticky top-0 z-50 glass-card border-x-0 border-t-0 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <a href="/" className="flex items-center gap-3 no-underline">
            <div className="flex items-center justify-center w-9 h-9 rounded-xl bg-gradient-to-tr from-violet-600 to-indigo-600 shadow-[0_0_12px_rgba(99,102,241,0.3)]">
              <Activity className="w-5 h-5 text-white" />
            </div>
            <div className="flex flex-col leading-tight">
              <span className="font-extrabold text-sm tracking-wider uppercase bg-clip-text text-transparent bg-gradient-to-r from-white to-gray-400">
                mcp-proxy
              </span>
              <span className="text-[10px] text-gray-500 hidden sm:block">
                高性能 MCP 网关代理
              </span>
            </div>
          </a>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center gap-1">
            {navItems.map(item => (
              <a
                key={item.id}
                href={item.href}
                className={`flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-bold uppercase tracking-wider transition-all duration-300 no-underline ${
                  activePage === item.id
                    ? "bg-violet-500/10 text-violet-400 border border-violet-500/20 shadow-[0_0_12px_rgba(99,102,241,0.08)]"
                    : "text-gray-400 hover:text-white hover:bg-white/5 border border-transparent hover:border-white/5"
                }`}
              >
                <item.icon className="w-4 h-4" />
                <span>{item.label}</span>
              </a>
            ))}
          </div>

          {/* Right Actions */}
          <div className="hidden md:flex items-center gap-3">
            <a
              href="https://github.com/ahaufox/mcp-gateway"
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1.5 px-3 py-1.5 bg-white/5 border border-white/10 hover:border-white/20 hover:bg-white/10 text-gray-400 hover:text-white rounded-xl text-xs font-semibold cursor-pointer transition-all duration-300 no-underline"
            >
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
              </svg>
              <span>GitHub</span>
            </a>
            {getAuthToken() && (
              <div className="flex items-center gap-1.5 px-3 py-1.5 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-xs text-emerald-400 font-semibold select-none">
                <ShieldCheck className="w-4 h-4" />
                <span>已授权</span>
              </div>
            )}
            {getAuthToken() && (
                <button
                  onClick={handleLogout}
                  className="flex items-center gap-1.5 px-3 py-1.5 bg-rose-500/5 border border-white/10 hover:border-rose-500/30 hover:bg-rose-500/10 text-gray-400 hover:text-rose-400 rounded-xl text-xs font-semibold cursor-pointer transition-all duration-300"
                >
                  <LogOut className="w-4 h-4" />
                  <span>退出</span>
                </button>
              )}
              <button
                onClick={toggleTheme}
                className="flex items-center gap-1.5 px-3 py-1.5 bg-white/5 border border-white/10 hover:border-violet-500/30 hover:bg-violet-500/10 text-gray-400 hover:text-violet-400 rounded-xl text-xs font-semibold cursor-pointer transition-all duration-300"
                title={theme === "dark" ? "切换到浅色模式" : "切换到深色模式"}
              >
                {theme === "dark" ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
                <span>{theme === "dark" ? "浅色" : "深色"}</span>
              </button>
          </div>

          {/* Mobile Menu Button */}
          <div className="flex md:hidden">
            <button
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              className="p-2 text-gray-400 hover:text-white transition-colors cursor-pointer"
            >
              {isMobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Dropdown Menu */}
      {isMobileMenuOpen && (
        <div className="md:hidden border-t border-white/5 bg-gray-950/95 backdrop-blur-xl px-4 py-4 space-y-3">
          {navItems.map(item => (
            <a
              key={item.id}
              href={item.href}
              onClick={() => setIsMobileMenuOpen(false)}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-300 no-underline ${
                activePage === item.id
                  ? "bg-violet-500/10 text-violet-400 border border-violet-500/10"
                  : "text-gray-400 hover:text-white"
              }`}
            >
              <item.icon className="w-5 h-5" />
              <span>{item.label}</span>
            </a>
          ))}

          <div className="border-t border-white/5 pt-3 space-y-3">
            <div className="flex items-center justify-between">
              <a
                href="https://github.com/ahaufox/mcp-gateway"
                target="_blank"
                rel="noreferrer"
                className="text-xs text-gray-500 flex items-center gap-1.5 px-3 py-1 hover:text-white transition-colors no-underline"
              >
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                </svg> GitHub 仓库
              </a>
            </div>
            <button
              onClick={() => {
                toggleTheme();
                setIsMobileMenuOpen(false);
              }}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-violet-500/10 border border-violet-500/20 text-violet-300 rounded-2xl text-xs font-semibold cursor-pointer"
            >
              {theme === "dark" ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
              <span>{theme === "dark" ? "切换到浅色模式" : "切换到深色模式"}</span>
            </button>
            {getAuthToken() && (
              <button
                onClick={() => {
                  handleLogout();
                  setIsMobileMenuOpen(false);
                }}
                className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-rose-500/10 border border-rose-500/20 text-rose-300 rounded-2xl text-xs font-semibold cursor-pointer"
              >
                <LogOut className="w-4 h-4" />
                <span>退出登录</span>
              </button>
            )}
          </div>
        </div>
      )}
    </nav>
  );
};