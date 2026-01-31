# -*- coding: utf-8 -*-
import os
import shutil
import hashlib
import subprocess
import sys
from pathlib import Path


def calculate_checksum(file_path, algorithm='sha256'):
    """计算文件的校验和"""
    hash_func = getattr(hashlib, algorithm)()
    with open(file_path, 'rb') as f:
        for chunk in iter(lambda: f.read(4096), b""):
            hash_func.update(chunk)
    return hash_func.hexdigest()


def clean_build_artifacts():
    """清理构建临时文件"""
    print("🧹 正在清理临时文件...")
    
    # 删除 build 文件夹
    if os.path.exists('build'):
        try:
            shutil.rmtree('build')
            print("   - 已删除 build 文件夹")
        except Exception as e:
            print(f"   ! 删除 build 文件夹失败: {e}")

    # 删除 spec 文件
    if os.path.exists('WT_Aimer_Voice.spec'):
        try:
            os.remove('WT_Aimer_Voice.spec')
            print("   - 已删除 spec 文件")
        except Exception as e:
            print(f"   ! 删除 spec 文件失败: {e}")


def load_dotenv(path=".env"):
    if os.path.exists(path):
        try:
            with open(path, "r", encoding="utf-8") as f:
                for line in f:
                    line = line.strip()
                    if not line or line.startswith("#") or "=" not in line:
                        continue
                    k, v = line.split("=", 1)
                    os.environ.setdefault(k.strip(), v.strip())
        except Exception as e:
            print(f"   ! 加载 .env 失败: {e}")


def build_exe():
    """执行打包任务"""
    print("🚀 开始打包程序...")
    
    # 确保 dist 目录存在 (PyInstaller 会自动创建，但为了保险)
    dist_dir = Path("dist")

    load_dotenv()
    
    # 在打包前，从打包环境的环境变量中读取加密salt和遥测url
    # 如果没有设置，则使用开发默认值
    salt = os.environ.get("TELEMETRY_SALT", "DEVELOPMENT_SALT")
    url = os.environ.get("REPORT_URL", "https://api.example.com/telemetry")
    
    # 生成临时的 app_secrets.py 供编译使用
    # 注意：该文件已被加入 .gitignore，不会被上传到 GitHub
    secrets_file = Path("app_secrets.py")
    with open(secrets_file, "w", encoding="utf-8") as f:
        f.write("# 由 build.py 自动生成 - 不要把它提交到github\n")
        f.write(f"TELEMETRY_SALT = {repr(salt)}\n")
        f.write(f"REPORT_URL = {repr(url)}\n")

    # PyInstaller 参数
    # --noconsole: 不显示控制台窗口
    # --onefile: 打包成单文件
    # --add-data: 添加资源文件 (Windows下用 ; 分隔)
    # --name: 指定生成的文件名
    # --icon: 指定图标
    
    cmd = [
        sys.executable, "-m", "PyInstaller",
        "--noconsole",
        "--onefile",
        "--hidden-import", "app_secrets",
        "--hidden-import", "requests",
        "--add-data", "web;web",
        "--name", "WT_Aimer_Voice",
        "--icon", "web/assets/logo.ico",
        "--clean",
        "main.py"
    ]
    
    print(f"执行命令: {' '.join(cmd)}")
    
    try:
        result = subprocess.run(cmd, check=True, shell=True, capture_output=True, text=True)
        print(result.stdout)
        print(result.stderr)
    except subprocess.CalledProcessError as e:
        print(f"[X] 打包失败！错误: {e}")
        print("--- PyInstaller stdout ---")
        print(e.stdout)
        print("--- PyInstaller stderr ---")
        print(e.stderr)
        import traceback
        traceback.print_exc()
        sys.exit(1)
    except Exception as e:
        print(f"[X] 打包失败！错误: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
    else:
        exe_path = Path("dist/WT_Aimer_Voice.exe")
        print(f"[OK] 打包成功！")
        print(f"输出文件: {exe_path}")
        return True


def main():
    # 1. 执行打包
    if not build_exe():
        return

    # 2. 生成校验文件
    exe_path = Path("dist/WT_Aimer_Voice.exe")
    if not exe_path.exists():
        print("❌ 未找到生成的 exe 文件！")
        return

    print("🔐 正在生成校验文件...")
    checksum = calculate_checksum(exe_path, 'sha256')
    checksum_file = dist_dir = Path("dist/checksum.txt")
    
    with open(checksum_file, 'w', encoding='utf-8') as f:
        f.write(f"File: {exe_path.name}\n")
        f.write(f"SHA256: {checksum}\n")
        f.write(f"Date: {os.popen('date /t').read().strip()} {os.popen('time /t').read().strip()}\n")
    
    print(f"✅ 校验文件已生成: {checksum_file}")
    print(f"   SHA256: {checksum}")

    # 3. 清理临时文件
    clean_build_artifacts()
    
    print("\n🎉 所有任务完成！可执行文件位于 dist 目录。")


if __name__ == "__main__":
    main()
