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
      // 验证 Token，向后端请求配置，验证是否能成功解析且不返回 401
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
    } catch (err: any) { // eslint-disable-line @typescript-eslint/no-explicit-any
      if (err.response && err.response.status === 401) {
        setError("无效的 Token，请重新输入");
      } else {
        setError(err.response?.data?.message || "连接服务器失败，请确保代理服务已启动");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-950 bg-gradient-mesh flex items-center justify-center p-4">
      <div className="w-full max-w-md stagger-in">
        {/* LOGO */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-tr from-violet-600 to-indigo-600 shadow-[0_0_20px_rgba(99,102,241,0.3)] mb-4">
            <KeyRound className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-white via-gray-200 to-gray-400 tracking-tight">
            MCP Proxy 控制台
          </h1>
          <p className="text-sm text-gray-400 mt-2">
            请输入管理员授权的 Access Token 以继续
          </p>
        </div>

        {/* 登录卡片 */}
        <div className="glass-card rounded-3xl p-8 backdrop-blur-xl relative overflow-hidden">
          <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-violet-500/0 via-violet-500/40 to-violet-500/0" />
          
          <form onSubmit={handleVerify} className="space-y-6">
            <div>
              <label htmlFor="token" className="block text-xs font-semibold uppercase tracking-wider text-gray-400 mb-2">
                Access Token
              </label>
              <div className="relative">
                <input
                  type="password"
                  id="token"
                  value={token}
                  onChange={(e) => {
                    setToken(e.target.value);
                    setError("");
                  }}
                  placeholder="输入授权 Token"
                  className="w-full bg-white/5 border border-white/10 rounded-2xl py-3 px-4 text-white placeholder-gray-500 transition-all duration-300 focus:bg-white/10 focus:border-violet-500/50"
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
              className="w-full bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white font-medium py-3 px-4 rounded-2xl transition-all duration-300 flex items-center justify-center gap-2 cursor-pointer shadow-[0_4px_12px_rgba(99,102,241,0.2)] disabled:opacity-50"
            >
              {loading ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  <span>正在验证...</span>
                </>
              ) : (
                <>
                  <span>进入控制台</span>
                  <ArrowRight className="w-5 h-5" />
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
