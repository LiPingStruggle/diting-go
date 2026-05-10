#!/bin/bash
cd "$(dirname "$0")"
git remote add origin https://github.com/LiPingStruggle/diting-go.git 2>/dev/null
git push -u origin master
echo "✅ 完成！访问: https://github.com/LiPingStruggle/diting-go"
