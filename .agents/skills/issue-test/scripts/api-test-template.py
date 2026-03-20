#!/usr/bin/env python3
# API 测试脚本模板
# 使用方法: python3 api-test-template.py <base_url> <auth_endpoint> [test_config_file]
# 示例: python3 api-test-template.py "http://localhost:8081" "/api/v1/user/auth/dev-login/1"

import subprocess
import sys
import os
import json
import time
from datetime import datetime


# 性能监测相关变量
API_TIMES = {}
TOTAL_API_CALLS = 0
TOTAL_TIME = 0
PASSED_TESTS = 0
FAILED_TESTS = 0
BEARER_TOKEN = ""

# 日志文件
TIMESTAMP = datetime.now().strftime("%Y%m%d_%H%M%S")
LOG_DIR = "./test_logs"
API_LOG = ""
RESULT_LOG = ""


def init_test_env(base_url):
    global API_LOG, RESULT_LOG
    os.makedirs(LOG_DIR, exist_ok=True)
    API_LOG = f"{LOG_DIR}/api_request_log_{TIMESTAMP}.json"
    RESULT_LOG = f"{LOG_DIR}/test_result_{TIMESTAMP}.log"

    with open(API_LOG, "w") as f:
        json.dump({"requests": [], "summary": {}}, f)
    with open(RESULT_LOG, "w") as f:
        f.write(f"API Test Started: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")

    print("=========================================")
    print("   API 接口测试")
    print(f"   Base URL: {base_url}")
    print("=========================================\n")


def measure_api_call(api_name, method, endpoint, data=None, expected_status="200", base_url=""):
    global TOTAL_API_CALLS, TOTAL_TIME, PASSED_TESTS, FAILED_TESTS

    url = f"{base_url}{endpoint}"
    start_time = time.time()

    # 构建curl命令
    curl_cmd = f"curl -s -w '\\n%{{http_code}}' -X {method} '{url}'"
    curl_cmd += " -H 'Content-Type: application/json'"

    if BEARER_TOKEN:
        curl_cmd += f" -H 'Authorization: Bearer {BEARER_TOKEN}'"

    if data and data != "null":
        curl_cmd += f" -d '{data}'"

    # 执行请求
    result = subprocess.run(curl_cmd, shell=True, capture_output=True, text=True)
    response = result.stdout.strip()
    end_time = time.time()
    duration = int((end_time - start_time) * 1000)

    # 分离响应体和状态码
    lines = response.rsplit("\n", 1)
    status_code = lines[-1] if len(lines) > 1 else "000"
    body = lines[0] if len(lines) > 1 else response

    # 性能评估
    if duration < 200:
        perf_status = "🟢 优秀"
    elif duration < 500:
        perf_status = "🟡 良好"
    elif duration < 1000:
        perf_status = "🟠 一般"
    else:
        perf_status = "🔴 慢"

    # 验证状态码
    if status_code == expected_status:
        test_result = "✅ PASS"
        PASSED_TESTS += 1
    else:
        test_result = f"❌ FAIL (期望: {expected_status}, 实际: {status_code})"
        FAILED_TESTS += 1

    # 显示结果
    print(f"\n----------------------------------------")
    print(f"[{api_name}] {method} {endpoint}")
    print(f"⏱️  响应时间: {duration}ms {perf_status}")
    print(f"状态码: {status_code} | 结果: {test_result}")
    print(f"\n返回结果:")

    # 尝试格式化 JSON
    try:
        parsed = json.loads(body)
        print(json.dumps(parsed, indent=2, ensure_ascii=False))
    except Exception:
        print(body)

    # 记录性能数据
    API_TIMES[api_name] = duration
    TOTAL_API_CALLS += 1
    TOTAL_TIME += duration

    # 记录到日志
    with open(RESULT_LOG, "a") as f:
        f.write(f"[{api_name}] {method} {endpoint} - {status_code} - {duration}ms\n")


def get_auth_token(base_url, auth_endpoint):
    global BEARER_TOKEN
    print("正在获取认证令牌...")
    result = subprocess.run(f"curl -s '{base_url}{auth_endpoint}'", shell=True, capture_output=True, text=True)
    try:
        data = json.loads(result.stdout)
        token = data.get("data") or data.get("token") or data.get("access_token")
        if token and token != "null":
            BEARER_TOKEN = str(token)
            print(f"Token获取成功: {BEARER_TOKEN[:50]}...\n")
            return True
    except Exception:
        pass
    print("⚠️  未能获取Token，将以无认证模式运行\n")
    return False


def generate_report():
    print(f"\n=========================================")
    print(f"   性能监测报告")
    print(f"=========================================")

    avg_time = TOTAL_TIME // TOTAL_API_CALLS if TOTAL_API_CALLS > 0 else 0

    print(f"\n各API响应时间统计:")
    for api_name, duration in API_TIMES.items():
        if duration < 200:
            status = "🟢 优秀"
        elif duration < 500:
            status = "🟡 良好"
        elif duration < 1000:
            status = "🟠 一般"
        else:
            status = "🔴 慢"
        print(f"  {api_name:<30} {duration:>6}ms {status}")

    print(f"\n----------------------------------------")
    print(f"测试统计:")
    print(f"  总API调用次数: {TOTAL_API_CALLS}")
    print(f"  通过: {PASSED_TESTS}")
    print(f"  失败: {FAILED_TESTS}")
    print(f"  总响应时间: {TOTAL_TIME}ms")
    print(f"  平均响应时间: {avg_time}ms")

    print(f"\n----------------------------------------")
    print(f"性能建议:")
    if avg_time < 200:
        print(f"  ✅ 所有API响应速度优秀（< 200ms）")
    elif avg_time < 500:
        print(f"  ⚠️  整体响应速度良好，建议监控大请求")
    elif avg_time < 1000:
        print(f"  ⚠️  整体响应速度一般，建议优化")
    else:
        print(f"  ❌ 响应速度较慢，急需优化")

    print(f"\n=========================================")
    if FAILED_TESTS == 0:
        print(f"✅ 所有测试通过！")
    else:
        print(f"❌ 有 {FAILED_TESTS} 个测试失败")
    print(f"测试完成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"日志文件: {RESULT_LOG}")
    print(f"=========================================")

    with open(RESULT_LOG, "a") as f:
        f.write(f"\n=========================================\n")
        f.write(f"总计: {TOTAL_API_CALLS} 调用, {PASSED_TESTS} 通过, {FAILED_TESTS} 失败\n")
        f.write(f"平均响应时间: {avg_time}ms\n")
        f.write(f"测试完成: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")


def run_example_tests(base_url):
    print("运行示例测试...")
    print("请在实际使用时替换为具体的API测试用例\n")
    # 示例:
    # measure_api_call("get_users", "GET", "/api/v1/users", base_url=base_url)
    # measure_api_call("create_user", "POST", "/api/v1/users", '{"name":"test"}', "201", base_url)


def main():
    base_url = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8081"
    auth_endpoint = sys.argv[2] if len(sys.argv) > 2 else "/api/v1/user/auth/dev-login/1"
    test_config = sys.argv[3] if len(sys.argv) > 3 else ""

    init_test_env(base_url)
    get_auth_token(base_url, auth_endpoint)

    if test_config and os.path.isfile(test_config):
        print(f"加载测试配置: {test_config}\n")
        # 动态加载配置文件中的测试用例
        import importlib.util
        spec = importlib.util.spec_from_file_location("test_config", test_config)
        module = importlib.util.module_from_spec(spec)
        module.measure_api_call = lambda *args, **kwargs: measure_api_call(*args, base_url=base_url, **kwargs)
        module.BEARER_TOKEN = BEARER_TOKEN
        spec.loader.exec_module(module)
        if hasattr(module, "run_tests"):
            module.run_tests()
    else:
        run_example_tests(base_url)

    generate_report()
    sys.exit(0 if FAILED_TESTS == 0 else 1)


if __name__ == "__main__":
    main()
