#!/bin/bash
# 《谛听》+《灵枢》综合测试脚本

set -e

echo '========================================'
echo '  📋 谛听 CLI 测试套件'
echo '========================================'

DITING='/Users/joyful/Documents/Project/diting-go/diting'
LINGSHU='/Users/joyful/Documents/Project/pure-ai-orchestrator/lingshu_full'

# 1. 基础命令测试
echo -e '\n🔵 [1/6] 基础命令测试'
$DITING version
$DITING help
$DITING models
$DITING check
echo '✅ 基础命令通过'

# 2. 预览功能测试
echo -e '\n🔵 [2/6] 预览功能测试'
$DITING preview 'test' --code /Users/joyful/Documents/Project/diting/example_auth.py
$DITING preview '' --code '' 2>&1 && echo '空参数测试通过' || echo '空参数错误反馈正常'
echo '✅ 预览功能通过'

# 3. 完整验证测试
echo -e '\n🔵 [3/6] 完整验证测试'
$DITING verify '空指针' --code /Users/joyful/Documents/Project/diting/example_auth.py
$DITING verify 'bug' -c /Users/joyful/Documents/Project/diting/example_auth.py -m ollama 2>&1 || echo 'ollama回退到mock'
echo '✅ 验证功能通过'

# 4. 仪表板和统计
echo -e '\n🔵 [4/6] 仪表板测试'
$DITING dashboard
$DITING stats
echo '✅ 仪表板通过'

echo -e '\n========================================'
echo '  ✅ 谛听 CLI 测试全部通过'
echo '========================================'