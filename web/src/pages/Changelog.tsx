import React from "react";
import { useTheme } from "../context/ThemeContext";
import { History, GitCommit, Calendar, Tag, Sparkles, Code2 } from "lucide-react";

interface ChangelogItem {
  version: string;
  date: string;
  colorClass: string;
  dotColor: string;
  items: string[];
}

const changelogData: ChangelogItem[] = [
  {
    version: "控制台架构重构 (SPA)",
    date: "2026-06-03",
    colorClass: "bg-fuchsia-500/10 text-fuchsia-400 border-fuchsia-500/20",
    dotColor: "bg-fuchsia-500",
    items: [
      "控制台全重构: 使用 React 19 + TypeScript + Vite + Tailwind CSS 重构为前后端分离架构，大幅度优化交互体验，且打包资产自动由 Go 二进制文件嵌入。",
      "鉴权机制升级: 废除多余的额外用户名密码配置，直接复用 Proxy 已有的全局 authTokens 进行拦截鉴权，前端自动管理 Token 状态，保障配置接口与监控接口的安全。",
      "UI 深度优化: 全站覆盖深色模式与 Glassmorphism（磨砂玻璃卡片）美学设计，增加卡片缩放微动画、错误日志抽屉及可视化转换器选择器。"
    ]
  },
  {
    version: "抖音 MCP 服务集成",
    date: "2026-02-14",
    colorClass: "bg-violet-500/10 text-violet-400 border-violet-500/20",
    dotColor: "bg-violet-500",
    items: [
      "新增服务: 集成 douyin-mcp 抖音视频解析服务，支持无水印下载、图文作品下载、AI 语音文案提取等功能。",
      "编排扩展: docker-compose 新增 douyin-mcp 与 jules-mcp-server 容器编排，支持健康检查与自动依赖启动。",
      "协议统一: Jules 和 Douyin 服务统一采用 SSE 传输协议，修复 Docker 内部网络通信异常。",
      "文档同步: README 新增已集成 MCP Server 一览表，完善项目结构描述。"
    ]
  },
  {
    version: "架构重构与监控增强",
    date: "2026-02-13",
    colorClass: "bg-blue-500/10 text-blue-400 border-blue-500/20",
    dotColor: "bg-blue-500",
    items: [
      "状态监控: Dashboard 新增 MCP 服务健康状态实时显示（连接/失败/不健康）及错误详情展示。",
      "标准架构: 重构项目目录为标准 Go 布局（cmd, internal），提升代码可维护性。",
      "环境隔离: 支持 AUTH_TOKENS 和 MCP_BASE_URL 环境变量注入，避免敏感配置硬编码。",
      "权限控制: 优化 Token 解析逻辑，支持通过环境变量注入逗号分隔的多 Token 列表。"
    ]
  },
  {
    version: "界面重构与汉化",
    date: "2026-02-12",
    colorClass: "bg-indigo-500/10 text-indigo-400 border-indigo-500/20",
    dotColor: "bg-indigo-500",
    items: [
      "功能增强: 配置转换器新增 Antigravity 格式支持，一键生成专用配置。",
      "视觉升级: 采用全新的 Glassmorphism 设计风格，统一 Dashboard 与工具页面视觉。",
      "体验优化: 完成全站中文化，精简页脚信息，从 docs 目录迁移至模板引擎渲染。"
    ]
  },
  {
    version: "初始化发布",
    date: "2024-03-20",
    colorClass: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    dotColor: "bg-emerald-500",
    items: [
      "核心发布: 发布核心 MCP 代理功能，支持 SSE 与 Streamable HTTP 传输协议。"
    ]
  }
];

export const Changelog: React.FC = () => {
  const { theme } = useTheme();
  const totalReleases = changelogData.length;
  const totalChanges = changelogData.reduce((sum, item) => sum + item.items.length, 0);
  const firstReleaseDate = changelogData[changelogData.length - 1]?.date;
  const latestReleaseDate = changelogData[0]?.date;

  return (
    <div className="space-y-8 stagger-in">
      {/* 头部 */}
      <div>
        <h1 className={`text-2xl font-bold tracking-tight flex items-center gap-2 ${theme === "dark" ? "text-white" : "text-gray-900"}`}>
          <History className="w-6 h-6 text-violet-500" />
          <span>更新日志</span>
        </h1>
        <p className={`text-sm mt-1 ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
          追踪 mcp-proxy 的版本迭代历程与核心技术演进
        </p>
      </div>

      {/* 版本概览统计卡片 */}
      <section className="grid grid-cols-2 md:grid-cols-4 gap-3">
        {[
          { icon: Tag, label: "累计版本", value: totalReleases, color: "violet", unit: "个" },
          { icon: Code2, label: "变更条目", value: totalChanges, color: "emerald", unit: "项" },
          { icon: Sparkles, label: "最新发布", value: latestReleaseDate, color: "indigo", unit: "", isText: true },
          { icon: Calendar, label: "起始时间", value: firstReleaseDate, color: "amber", unit: "", isText: true }
        ].map((item, idx) => {
          const colorMap: Record<string, { bg: string; text: string; border: string }> = {
            violet: { bg: "bg-violet-500/10", text: "text-violet-400", border: "border-violet-500/20" },
            emerald: { bg: "bg-emerald-500/10", text: "text-emerald-400", border: "border-emerald-500/20" },
            indigo: { bg: "bg-indigo-500/10", text: "text-indigo-400", border: "border-indigo-500/20" },
            amber: { bg: "bg-amber-500/10", text: "text-amber-400", border: "border-amber-500/20" }
          };
          const c = colorMap[item.color];
          return (
            <div key={idx} className={`glass-card rounded-2xl p-4 border flex items-center gap-3 ${theme === "dark" ? "border-white/10" : "border-gray-200"}`}>
              <div className={`p-2 rounded-xl ${c.bg} border ${c.border} shrink-0`}>
                <item.icon className={`w-5 h-5 ${c.text}`} />
              </div>
              <div className="min-w-0">
                <div className={`text-lg font-extrabold truncate ${item.isText ? "text-base" : ""} ${theme === "dark" ? "text-white" : "text-gray-900"}`}>
                  {item.value}{!item.isText && item.unit && <span className={`text-sm font-bold ${c.text} ml-0.5`}>{item.unit}</span>}
                </div>
                <div className={`text-[10px] font-bold uppercase tracking-wider ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>{item.label}</div>
              </div>
            </div>
          );
        })}
      </section>

      {/* 垂直时间线布局 */}
      <div className="relative pl-6 md:pl-8 space-y-12">
        <div className="timeline-line" />

        {changelogData.map((version, index) => (
          <div key={index} className="relative stagger-in" style={{ animationDelay: `${index * 0.1}s` }}>
            {/* 时间点 */}
            <span className={`absolute -left-[27px] md:-left-[31px] top-1.5 flex h-4 w-4 rounded-full border-4 border-gray-950 ${version.dotColor} shadow-[0_0_8px_rgba(255,255,255,0.15)]`} />

            {/* 卡片 */}
            <div className="glass-card rounded-3xl p-6 md:p-8 hover:shadow-[0_4px_24px_rgba(99,102,241,0.02)]">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-6">
                <h2 className={`text-lg font-bold tracking-tight flex items-center gap-2 ${theme === "dark" ? "text-white" : "text-gray-900"}`}>
                  <GitCommit className="w-4 h-4 text-violet-400" />
                  <span>{version.version}</span>
                </h2>
                <span className={`px-3 py-1 rounded-xl text-xs font-bold border self-start sm:self-center ${version.colorClass}`}>
                  {version.date}
                </span>
              </div>

              <ul className="space-y-3.5">
                {version.items.map((bullet, bidx) => (
                  <li key={bidx} className={`flex items-start text-xs leading-relaxed ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
                    <span className={`w-1.5 h-1.5 rounded-full ${version.dotColor} shrink-0 mt-1.5 mr-3 opacity-60`} />
                    <span>{bullet}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
