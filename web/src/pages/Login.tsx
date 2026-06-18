import React, { useState } from "react";
import { setAuthToken } from "../utils/api";
import axios from "axios";
import { KeyRound, ShieldAlert, ArrowRight, Loader2 } from "lucide-react";

interface LoginProps {
  onLoginSuccess: () => void;
}

export const Login: React.FC<LoginProps> = ({ onLoginSuccess }) => {
  const [token, setToken] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token.trim()) {
      setError("请输入访问 Token");
      return;
    }

    setLoading(true);
    setError("");

    try {
      const response = await axios.get("/api/config", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.status === 200) {
        setAuthToken(token);
        onLoginSuccess();
      } else {
        setError("Token 校验失败，请重试");
      }
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        if (err.response && err.response.status === 401) {
          setError("无效的 Token，请重新输入");
        } else {
          setError(err.response?.data?.message || "连接服务器失败，请确保代理服务已启动");
        }
      } else {
        setError("连接服务器失败，请确保代理服务已启动");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-950 bg-gradient-mesh flex items-center justify-center p-4 relative overflow-hidden">
      {/* 漂浮光斑 */}
      <div className="absolute top-[-100px] left-[-100px] w-[400px] h-[400px] rounded-full blur-[80px] opacity-[0.22] pointer-events-none"
        style={{ background: "radial-gradient(circle, #6366F1 0%, transparent 70%)", animation: "float 12s ease-in-out infinite" }} />
      <div className="absolute top-[30%] right-[-150px] w-[500px] h-[500px] rounded-full blur-[80px] opacity-[0.22] pointer-events-none"
        style={{ background: "radial-gradient(circle, #A855F7 0%, transparent 70%)", animation: "float 12s ease-in-out infinite", animationDelay: "-4s" }} />
      <div className="absolute bottom-[-100px] left-[30%] w-[350px] h-[350px] rounded-full blur-[80px] opacity-[0.22] pointer-events-none"
        style={{ background: "radial-gradient(circle, #3B82F6 0%, transparent 70%)", animation: "float 12s ease-in-out infinite", animationDelay: "-8s" }} />

      <div className="relative z-10 w-full max-w-md stagger-in">
        {/* 品牌区 */}
        <div className="flex flex-col items-center mb-10">
          <div className="relative w-20 h-20 mb-6">
            <div className="absolute inset-0 bg-gradient-to-br from-violet-500 to-indigo-600 rounded-2xl rotate-6 blur-xl opacity-60" />
            <div className="relative w-full h-full bg-gradient-to-br from-violet-500/90 to-indigo-600/90 rounded-2xl flex items-center justify-center glass">
              <span className="mono text-white text-2xl font-bold tracking-tighter">M</span>
            </div>
          </div>
          <h1 className="text-4xl font-bold tracking-tight text-gradient mb-2">
            mcp-proxy
          </h1>
          <p className="text-[13px] text-violet-200/60 tracking-wide">
            Model Context Protocol Gateway
          </p>
        </div>

        {/* 登录表单卡片 */}
        <div className="glass rounded-3xl p-8 tilt-3d">
          <div className="mb-6">
            <h2 className="text-[22px] font-semibold text-white mb-1.5">欢迎回来</h2>
            <p className="text-[13px] text-violet-200/60">输入访问令牌以进入控制台</p>
          </div>

          <form onSubmit={handleVerify} className="space-y-4">
            <div>
              <label className="block text-[11px] font-medium uppercase tracking-[0.15em] text-violet-200/70 mb-2">
                Access Token
              </label>
              <div className="relative">
                <KeyRound className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-violet-300/50" />
                <input
                  type="password"
                  placeholder="••••••••••••••••"
                  value={token}
                  onChange={(e) => {
                    setToken(e.target.value);
                    setError("");
                  }}
                  className="input-vp w-full h-12 pl-11 pr-4 rounded-xl text-[14px] text-white placeholder-violet-300/30 mono"
                  disabled={loading}
                />
              </div>
            </div>

            {error && (
              <div className="flex items-center gap-3 bg-rose-500/10 border border-rose-500/20 text-rose-300 text-sm py-3 px-4 rounded-2xl animate-shake">
                <ShieldAlert className="w-5 h-5 shrink-0 text-rose-400" />
                <span>{error}</span>
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="btn-glow w-full h-12 rounded-xl text-white text-[14px] font-semibold mt-2 flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span>正在验证...</span>
                </>
              ) : (
                <>
                  <span>验证并进入</span>
                  <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </form>

          {/* 分割线 */}
          <div className="flex items-center gap-3 my-6">
            <div className="flex-1 h-px bg-gradient-to-r from-transparent to-white/10" />
            <span className="text-[10px] uppercase tracking-[0.2em] text-violet-200/40">安全连接</span>
            <div className="flex-1 h-px bg-gradient-to-l from-transparent to-white/10" />
          </div>

          {/* 状态指示 */}
          <div className="flex items-center justify-center gap-6 text-[11px] text-violet-200/50">
            <div className="flex items-center gap-1.5">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-lg shadow-emerald-400/50" />
              <span>TLS 加密</span>
            </div>
            <div className="flex items-center gap-1.5">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-lg shadow-emerald-400/50" />
              <span>本地代理</span>
            </div>
            <div className="flex items-center gap-1.5">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-lg shadow-emerald-400/50" />
              <span>v1.0.0</span>
            </div>
          </div>
        </div>

        {/* 底部链接 */}
        <p className="text-center mt-8 text-[11px] text-violet-200/40">
          没有令牌？
          <a href="#" className="text-violet-300 hover:text-white transition-colors ml-1">
            查看文档
          </a>
        </p>
      </div>
    </div>
  );
};