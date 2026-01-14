/**
 * =============================================================================
 * AIMER UI CONSTITUTION (Aimer 界面开发宪法) - V1.0
 * =============================================================================
 * * 【前言 / PREAMBLE】
 * 本文件是 Aimer UI 系统的最高法则。任何样式开发必须遵循以下条款。
 * 违背本宪法将导致：样式无法动态切换、分享码失效、以及维护者的愤怒。
 * * 【第一条：HEX 禁令 / THE PROHIBITION OF HEX】
 * CSS 文件 (.css) 中严禁出现任何硬编码的 HEX 颜色代码 (如 #FFFFFF, #333)。
 * 唯一允许的例外是：纯黑 (#000000)、纯白 (#FFFFFF) 用于遮罩层，以及 transparent。
 * 所有的颜色必须通过 var(--variable-name) 引用。
 * * 【第二条：注册义务 / REGISTRATION OBLIGATION】
 * 任何一个新的 UI 元素若需要独立变色，必须先在本文件的 THEME_SCHEMA 中注册。
 * 未注册的变量名视为“非法黑户”，生成分享码时将被丢弃。
 * * 【第三条：语义化命名 / SEMANTIC NAMING】
 * 变量名必须描述“它是谁”，而不是“它是什么颜色”。
 * - 正确 (√): --icon-trash-hover (垃圾桶悬停色)
 * - 错误 (×): --red-color (红色)
 * * 【第四条：短码映射 / SHORTCODE MAPPING】
 * 为了缩短分享口令，每个变量必须拥有一个 2-4 位的唯一短码 (Key)。
 * =============================================================================
 */

// 1. 变量映射字典 (Schema)
// 左侧是短码 (用于生成分享口令)，右侧是 CSS 变量名
const THEME_SCHEMA = {
    // --- [A] 全局基础 (Global) ---
    "p":   "--primary",           // 主题色
    "ph":  "--primary-hover",     // 主题色悬停
    "bg":  "--bg-body",           // 窗口背景
    "cbg": "--bg-card",           // 卡片背景
    "tm":  "--text-main",         // 主文字色
    "ts":  "--text-sec",          // 次文字色 (灰色)
    "bd":  "--border-color",      // 边框颜色

    // --- [B] 导航栏 (Nav) ---
    "nbg": "--nav-bg",            // 导航栏背景
    "ni":  "--nav-item-text",     // 导航默认文字
    "nih": "--nav-item-hover-bg", // 导航悬停背景
    "nia": "--nav-item-active",   // 导航激活文字
    "nibt":"--nav-item-active-bg",// 导航激活背景

    // --- [C] 状态指示 (Status) ---
    "stw": "--status-waiting",    // 黄灯
    "sts": "--status-success",    // 绿灯
    "ste": "--status-error",      // 红灯
    "sti": "--status-icon-def",   // 图标默认灰

    // --- [D] 语音包卡片 (Mod Card) ---
    "mct": "--mod-card-title",    // 标题文字
    "mcv": "--mod-ver-bg",        // 版本号背景
    "mcvt":"--mod-ver-text",      // 版本号文字
    "mca": "--mod-author-text",   // 作者文字
    
    // --- [E] 交互按钮 (Actions) ---
    "act": "--action-trash",      // 垃圾桶默认
    "acth":"--action-trash-hover",// 垃圾桶悬停(红)
    "acr": "--action-refresh",    // 刷新按钮文字
    "acrh":"--action-refresh-bg", // 刷新按钮悬停

    // --- [F] 外部链接 (Links) ---
    "lkb": "--link-bili-normal",  // B站
    "lkbh":"--link-bili-hover",
    "lkw": "--link-wt-normal",    // 官网
    "lkwh":"--link-wt-hover",
    "lkv": "--link-vid-normal",   // 视频
    "lkvh":"--link-vid-hover",

    // --- [G] 标签颜色 (Tags) ---
    "ttbg": "--tag-tank-bg",      "tttx": "--tag-tank-text",
    "tabg": "--tag-air-bg",       "tatx": "--tag-air-text",
    "tnbg": "--tag-naval-bg",     "tntx": "--tag-naval-text",
    "trbg": "--tag-radio-bg",     "trtx": "--tag-radio-text",

    // --- [H] 滚动条 (Scrollbar) ---
    "sb":  "--scrollbar-thumb",    // 滑块颜色
    "sbh": "--scrollbar-thumb-hover", // 滑块悬停

    // --- [I] 输入框 (Inputs) ---
    "inp":  "--input-bg",          // 输入框背景
    "inpb": "--input-border",      // 输入框边框

    // --- [J] 模态框 (Modal) ---
    "modl": "--modal-overlay-bg",  // 遮罩层颜色

    // --- [K] 窗口控制 (Window Controls) ---
    "wcb":  "--win-btn-hover-bg"   // 窗口按钮悬停背景
};

// =============================================================================
// 默认主题定义 (SYSTEM DEFAULTS)
// =============================================================================

// 1. 默认明亮 (Light)
const DEFAULT_LIGHT = {
    "--primary": "#FF9900",
    "--primary-hover": "#e68a00",
    "--bg-body": "#F5F7FA",
    "--bg-card": "#FFFFFF",
    "--text-main": "#2C3E50",
    "--text-sec": "#7F8C8D",
    "--border-color": "#E2E8F0",

    "--nav-bg": "#FFFFFF",
    "--nav-item-text": "#7F8C8D",
    "--nav-item-hover-bg": "rgba(0, 0, 0, 0.05)",
    "--nav-item-active": "#FF9900",
    "--nav-item-active-bg": "rgba(255, 153, 0, 0.1)",

    "--status-waiting": "#F59E0B",
    "--status-success": "#10B981",
    "--status-error": "#EF4444",
    "--status-icon-def": "#E2E8F0",

    "--mod-card-title": "#2C3E50",
    "--mod-ver-bg": "rgba(255,153,0,0.1)",
    "--mod-ver-text": "#FF9900",
    "--mod-author-text": "#7F8C8D",

    "--action-trash": "#2C3E50",
    "--action-trash-hover": "#EF4444",
    "--action-refresh": "#2C3E50",
    "--action-refresh-bg": "#2C3E50",

    "--link-bili-normal": "#23ade5",
    "--link-bili-hover": "#23ade5",
    "--link-wt-normal": "#2C3E50",
    "--link-wt-hover": "#2C3E50",
    "--link-vid-normal": "#EF4444",
    "--link-vid-hover": "#EF4444",

    "--tag-tank-bg": "#DCFCE7", "--tag-tank-text": "#16A34A",
    "--tag-air-bg": "#F3F4F6", "--tag-air-text": "#4B5563",
    "--tag-naval-bg": "#E0F2FE", "--tag-naval-text": "#0284C7",
    "--tag-radio-bg": "#FEF9C3", "--tag-radio-text": "#CA8A04",

    "--scrollbar-thumb": "#CCCCCC",
    "--scrollbar-thumb-hover": "#FF9900",
    "--input-bg": "#FFFFFF",
    "--input-border": "#E2E8F0",
    "--modal-overlay-bg": "rgba(0, 0, 0, 0.5)",
    "--win-btn-hover-bg": "rgba(0, 0, 0, 0.05)"
};

// 2. 默认深色 (Dark) - 锌色系专业深色模式 (Zinc Palette) 
const DEFAULT_DARK = { 
    "--primary": "#F59E0B", 
    "--primary-hover": "#FBBF24", 
    "--bg-body": "#18181B", 
    "--bg-card": "#27272A", 
    "--text-main": "#F4F4F5", 
    "--text-sec": "#A1A1AA", 
    "--border-color": "#3F3F46", 

    "--nav-bg": "#27272A", 
    "--nav-item-text": "#A1A1AA", 
    "--nav-item-hover-bg": "rgba(255, 255, 255, 0.08)", 
    "--nav-item-active": "#FBBF24", 
    "--nav-item-active-bg": "rgba(251, 191, 36, 0.15)", 

    "--status-waiting": "#F59E0B", 
    "--status-success": "#34D399", 
    "--status-error": "#F87171", 
    "--status-icon-def": "#52525B", 

    "--mod-card-title": "#F4F4F5", 
    "--mod-ver-bg": "rgba(251, 191, 36, 0.15)", 
    "--mod-ver-text": "#FBBF24", 
    "--mod-author-text": "#71717A", 

    "--action-trash": "#71717A", 
    "--action-trash-hover": "#EF4444", 
    "--action-refresh": "#F4F4F5", 
    "--action-refresh-bg": "#F59E0B", 

    "--link-bili-normal": "#38BDF8", 
    "--link-bili-hover": "#E0F2FE", 
    "--link-wt-normal": "#9CA3AF", 
    "--link-wt-hover": "#F3F4F6", 
    "--link-vid-normal": "#F87171", 
    "--link-vid-hover": "#FEF2F2", 

    "--tag-tank-bg": "#14532D", "--tag-tank-text": "#4ADE80", 
    "--tag-air-bg": "#374151", "--tag-air-text": "#D1D5DB", 
    "--tag-naval-bg": "#1E3A8A", "--tag-naval-text": "#60A5FA", 
    "--tag-radio-bg": "#422006", "--tag-radio-text": "#FACC15", 

    "--scrollbar-thumb": "#52525B", 
    "--scrollbar-thumb-hover": "#F59E0B", 
    "--input-bg": "#09090B", 
    "--input-border": "#3F3F46", 
    "--modal-overlay-bg": "rgba(0, 0, 0, 0.8)", 
    "--win-btn-hover-bg": "rgba(255, 255, 255, 0.1)" 
};