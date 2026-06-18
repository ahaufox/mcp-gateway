import React from "react";
import { useTheme } from "../context/ThemeContext";
import { Rss, BellPlus } from "lucide-react";

interface ChangelogEntry {
  version: string;
  date: string;
  subtitle: string;
  title: string;
  color: string;       // Tailwind color key: violet | cyan | emerald | amber | slate
  isLatest: boolean;
  items: string[];
}

const changelogData: ChangelogEntry[] = [
  {
    version: "v2.0.0",
    date: "2026-06-03",
    subtitle: "控制台架构重构 (SPA)",
    title: "全新 React 19 + TypeScript 控制台",
    color: "violet",
    isLatest: true,
    items: [
      "控制台全重构 — React 19 + TypeScript + Vite + Tailwind CSS，资产自动由 Go 二进制嵌入",
      "鉴权机制升级 — 复用 Proxy 全局 authTokens，前端自动管理 Token 状态",
      "UI 深度优化 — 全站深色模式与 Glassmorphism 设计，卡片缩放动画与可视化转换器",
    ],
  },
  {
    version: "v1.3.0",
    date: "2026-02-14",
    subtitle: "抖音 MCP 服务集成",
    title: "集成 douyin-mcp 视频解析",
    color: "cyan",
    isLatest: false,
    items: [
      "新增服务 — 支持无水印下载、图文作品下载、AI 语音文案提取",
      "协议统一 — Jules 和 Douyin 服务统一采用 SSE 传输协议",
      "编排扩展 — docker-compose 新增容器编排与健康检查",
    ],
  },
  {
    version: "v1.2.0",
    date: "2026-02-13",
    subtitle: "架构重构与监控增强",
    title: "实时健康监控与标准 Go 布局",
    color: "emerald",
    isLatest: false,
    items: [
      "状态监控 — Dashboard 新增 MCP 服务健康状态实时显示",
      "标准架构 — 重构项目目录为标准 Go 布局（cmd, internal）",
      "环境隔离 — 支持 AUTH_TOKENS 与 MCP_BASE_URL 注入",
    ],
  },
  {
    version: "v1.1.0",
    date: "2026-02-12",
    subtitle: "界面重构与汉化",
    title: "Glassmorphism 设计与全站汉化",
    color: "amber",
    isLatest: false,
    items: [
      "功能增强 — 配置转换器新增 Antigravity 格式支持",
      "视觉升级 — 全新 Glassmorphism 设计风格统一全站",
      "体验优化 — 完成全站中文化，精简页脚信息",
    ],
  },
  {
    version: "v1.0.0",
    date: "2024-03-20",
    subtitle: "初始化发布",
    title: "核心 MCP 代理首发",
    color: "slate",
    isLatest: false,
    items: [
      "核心发布 — 发布 MCP 代理功能，支持 SSE 与 Streamable HTTP 传输协议",
    ],
  },
];

// ── 颜色映射 ──
const colorMap: Record<string, {
  badgeBg: string; badgeText: string;
  dotBg: string;
  blurBg: string;
  statBlur: string; statLabel: string;
  itemDot: string;
}> = {
  violet: {
    badgeBg: "bg-gradient-to-br from-violet-500 to-indigo-600",
    badgeText: "text-white",
    dotBg: "bg-gradient-to-br from-violet-400 to-indigo-600",
    blurBg: "bg-violet-500/20",
    statBlur: "bg-violet-500/20",
    statLabel: "text-violet-300/60",
    itemDot: "bg-violet-400",
  },
  cyan: {
    badgeBg: "glass-strong",
    badgeText: "text-cyan-300",
    dotBg: "bg-gradient-to-br from-cyan-400 to-blue-600",
    blurBg: "bg-cyan-500/20",
    statBlur: "bg-cyan-500/20",
    statLabel: "text-cyan-300/60",
    itemDot: "bg-cyan-400",
  },
  emerald: {
    badgeBg: "glass-strong",
    badgeText: "text-emerald-300",
    dotBg: "bg-gradient-to-br from-emerald-400 to-cyan-600",
    blurBg: "bg-emerald-500/20",
    statBlur: "bg-emerald-500/20",
    statLabel: "text-emerald-300/60",
    itemDot: "bg-emerald-400",
  },
  amber: {
    badgeBg: "glass-strong",
    badgeText: "text-amber-300",
    dotBg: "bg-gradient-to-br from-amber-400 to-orange-600",
    blurBg: "bg-amber-500/20",
    statBlur: "bg-amber-500/20",
    statLabel: "text-amber-300/60",
    itemDot: "bg-amber-400/60",
  },
  slate: {
    badgeBg: "glass-strong",
    badgeText: "text-slate-300",
    dotBg: "bg-gradient-to-br from-slate-400 to-slate-600",
    blurBg: "bg-slate-500/20",
    statBlur: "bg-slate-500/20",
    statLabel: "text-slate-300/60",
    itemDot: "bg-slate-400",
  },
};

// ── 统计指标数据 ──
const stats = [
  {
    value: changelogData.length,
    unit: "个",
    label: "累计版本",
    color: "violet" as const,
  },
  {
    value: changelogData.reduce((s, e) => s + e.items.length, 0),
    unit: "项",
    label: "变更条目",
    color: "cyan" as const,
  },
  {
    value: changelogData[0].date,
    unit: "",
    label: "最新发布",
    color: "emerald" as const,
  },
  {
    value: changelogData[changelogData.length - 1].date,
    unit: "",
    label: "起始时间",
    color: "amber" as const,
  },
];

export const Changelog: React.FC = () => {
  const { theme } = useTheme();
  const isDark = theme === "dark";

  return (
    <div className="space-y-8 stagger-in max-w-5xl mx-auto">
      {/* ── 头部 ── */}
      <header className="flex items-center justify-between border-b pb-4"
        style={{ borderColor: isDark ? "rgba(255,255,255,0.05)" : "rgba(0,0,0,0.06)" }}>
        <div>
          <h1 className={`text-xl font-semibold tracking-tight ${isDark ? "text-gradient" : "text-gray-900"}`}>
            更新日志
          </h1>
          <p className={`text-[11px] mt-0.5 ${isDark ? "text-violet-200/50" : "text-gray-500"}`}>
            追踪 mcp-proxy 的版本迭代历程
          </p>
        </div>
        <div className="flex items-center gap-2">
          <a
            href="#"
            className={`h-9 px-3 rounded-lg text-xs flex items-center gap-1.5 no-underline transition-colors ${
              isDark
                ? "glass-strong text-violet-200/80 hover:text-white"
                : "bg-gray-100 border border-gray-200 text-gray-600 hover:text-gray-900"
            }`}
          >
            <Rss className="w-3.5 h-3.5" />
            <span>RSS</span>
          </a>
          <button
            className={`h-9 px-3 rounded-lg text-white text-xs font-semibold flex items-center gap-1.5 cursor-pointer ${
              isDark ? "btn-glow" : "bg-violet-600 hover:bg-violet-700 shadow-md"
            }`}
          >
            <BellPlus className="w-3.5 h-3.5" />
            <span>订阅更新</span>
          </button>
        </div>
      </header>

      {/* ── 统计指标卡片 ── */}
      <section className="grid grid-cols-2 md:grid-cols-4 gap-3">
        {stats.map((s, idx) => {
          const cm = colorMap[s.color];
          const isText = s.color === "emerald" || s.color === "amber";
          return (
            <div
              key={idx}
              className={`glass hover-lift rounded-2xl p-5 relative overflow-hidden ${
                isDark ? "" : "!bg-white/90 !border-gray-200"
              }`}
            >
              <div
                className={`absolute -top-8 -right-8 w-24 h-24 rounded-full blur-2xl ${cm.statBlur}`}
              />
              <div className="relative">
                <div
                  className={`text-[10px] font-semibold uppercase tracking-[0.15em] mb-3 ${cm.statLabel}`}
                >
                  {s.label}
                </div>
                <div className="flex items-baseline gap-1">
                  <span
                    className={`${
                      isText ? "text-[15px] font-semibold" : "text-[32px] font-bold leading-none"
                    } ${
                      isText
                        ? s.color === "emerald"
                          ? "text-emerald-300"
                          : "text-amber-300"
                        : isDark
                          ? "text-gradient"
                          : "text-gray-900"
                    } mono`}
                  >
                    {s.value}
                  </span>
                  {s.unit && (
                    <span className={`text-xs ${isDark ? "text-violet-300/40" : "text-gray-400"}`}>
                      {s.unit}
                    </span>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </section>

      {/* ── 时间线 ── */}
      <div className="relative pl-10">
        {/* 渐变主线 */}
        <div className="timeline-line-enhanced" />

        {changelogData.map((entry, index) => {
          const cm = colorMap[entry.color];
          const isOld = entry.color === "amber" || entry.color === "slate";
          const opacityClass = entry.color === "amber" ? "opacity-90" : entry.color === "slate" ? "opacity-80" : "";

          return (
            <div key={index} className={`relative ${index < changelogData.length - 1 ? "pb-10" : ""}`}>
              {/* 时间线圆点 */}
              <div className="absolute -left-10 top-2 flex items-center justify-center">
                <div
                  className={`${
                    entry.isLatest ? "w-5 h-5" : "w-4 h-4"
                  } rounded-full ${cm.dotBg} ${
                    entry.isLatest ? "timeline-dot pulse-dot-anim" : isOld ? "timeline-dot-old" : "timeline-dot"
                  }`}
                />
              </div>

              {/* 版本卡片 */}
              <div
                className={`glass hover-lift rounded-2xl p-7 relative overflow-hidden ${opacityClass} ${
                  isDark ? "" : "!bg-white/90 !border-gray-200"
                }`}
              >
                {/* 模糊光斑（仅最新版本） */}
                {entry.isLatest && (
                  <div className="absolute -top-20 -right-20 w-60 h-60 bg-violet-500/20 rounded-full blur-3xl pointer-events-none" />
                )}

                {/* 模糊光斑（其他版本） */}
                {!entry.isLatest && (
                  <div className={`absolute -top-12 -right-12 w-40 h-40 rounded-full blur-2xl pointer-events-none ${cm.blurBg}`} />
                )}

                <div className="relative">
                  {/* 卡片头部 */}
                  <div
                    className="flex items-start justify-between mb-5 pb-5 border-b"
                    style={{ borderColor: isDark ? "rgba(255,255,255,0.05)" : "rgba(0,0,0,0.06)" }}
                  >
                    <div>
                      <div className="flex items-center gap-2.5 mb-2">
                        <span
                          className={`mono text-[10px] px-2 h-5 inline-flex items-center rounded-md ${cm.badgeBg} ${cm.badgeText} uppercase tracking-wider font-semibold`}
                        >
                          {entry.version}
                        </span>
                        <span className={`w-1 h-1 rounded-full ${isDark ? "bg-violet-300/40" : "bg-gray-300"}`} />
                        <span className={`text-[11px] ${isDark ? "text-violet-300/70" : "text-gray-500"}`}>
                          {entry.subtitle}
                        </span>
                        {entry.isLatest && (
                          <span className="mono text-[10px] px-1.5 py-0.5 rounded bg-violet-500/15 text-violet-300 uppercase tracking-wider">
                            Latest
                          </span>
                        )}
                      </div>
                      <h3 className={`text-xl font-semibold tracking-tight ${isDark ? "text-white" : "text-gray-900"}`}>
                        {entry.title}
                      </h3>
                    </div>
                    <span className={`mono text-[11px] shrink-0 ${isDark ? "text-violet-300/50" : "text-gray-400"}`}>
                      {entry.date}
                    </span>
                  </div>

                  {/* 变更条目列表 */}
                  <ul className="space-y-3.5">
                    {entry.items.map((item, iidx) => (
                      <li
                        key={iidx}
                        className={`flex items-start gap-3 text-[13px] leading-relaxed ${
                          isDark ? "text-violet-200/80" : "text-gray-600"
                        }`}
                      >
                        {entry.isLatest ? (
                          <div className="w-5 h-5 rounded-md bg-gradient-to-br from-violet-500/30 to-indigo-500/30 border border-violet-400/30 flex items-center justify-center shrink-0 mt-0.5">
                            <svg className="w-3 h-3 text-violet-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
                              <polyline points="20 6 9 17 4 12" />
                            </svg>
                          </div>
                        ) : (
                          <div className={`w-1.5 h-1.5 rounded-full ${cm.itemDot} mt-2 shrink-0`} />
                        )}
                        <div>
                          {entry.isLatest ? (
                            (() => {
                              const colonIdx = item.indexOf("—");
                              if (colonIdx > 0) {
                                return (
                                  <>
                                    <span className={`font-semibold ${isDark ? "text-white" : "text-gray-900"}`}>
                                      {item.slice(0, colonIdx).trim()}
                                    </span>
                                    <span>{item.slice(colonIdx)}</span>
                                  </>
                                );
                              }
                              return <span>{item}</span>;
                            })()
                          ) : (
                            (() => {
                              const colonIdx = item.indexOf("—");
                              if (colonIdx > 0) {
                                return (
                                  <>
                                    <span className={`font-semibold ${isDark ? "text-white" : "text-gray-900"}`}>
                                      {item.slice(0, colonIdx).trim()}
                                    </span>
                                    <span>{item.slice(colonIdx)}</span>
                                  </>
                                );
                              }
                              return <span>{item}</span>;
                            })()
                          )}
                        </div>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};