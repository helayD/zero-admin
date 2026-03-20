#!/usr/bin/env python3
# API 测试配置模板
# 复制此文件并根据 issue 内容填充测试用例
# 使用方法:
#   1. cp test-config-template.py test-config-<issue>.py
#   2. 编辑 run_tests 函数添加测试用例
#   3. python3 api-test-template.py "http://localhost:8081" "/api/v1/auth" "./test-config-<issue>.py"

# 可选: 覆盖默认配置
# BASE_URL = "http://localhost:8081"
# AUTH_ENDPOINT = "/api/v1/user/auth/dev-login/1"

# 测试数据变量 (根据 issue 内容填充)
# TEST_USER_ID = 1
# TEST_RESOURCE_ID = 1


def run_tests():
    # ========================================
    # 在此添加 API 测试用例
    # 使用 measure_api_call 函数:
    #   measure_api_call(<名称>, <方法>, <端点>, <数据>, <期望状态码>)
    # ========================================

    # 示例: GET 请求
    # measure_api_call("get_list", "GET", "/api/v1/resources")

    # 示例: GET 带参数
    # measure_api_call("get_detail", "GET", "/api/v1/resources/1")

    # 示例: POST 创建
    # measure_api_call("create_resource", "POST", "/api/v1/resources",
    #     '{"name":"test","value":123}', "201")

    # 示例: PUT 更新
    # measure_api_call("update_resource", "PUT", "/api/v1/resources/1",
    #     '{"name":"updated"}')

    # 示例: DELETE 删除
    # measure_api_call("delete_resource", "DELETE", "/api/v1/resources/1", None, "204")

    print("请编辑此文件添加测试用例")
