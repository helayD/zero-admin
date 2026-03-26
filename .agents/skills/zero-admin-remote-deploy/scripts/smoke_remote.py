#!/usr/bin/env python3
import argparse
import json
import sys
import urllib.error
import urllib.request
from typing import Any, Dict, Optional, Tuple


class SmokeCheckError(Exception):
    pass


def request_json(
    base_url: str,
    path: str,
    method: str = "GET",
    headers: Optional[Dict[str, str]] = None,
    payload: Optional[Dict[str, Any]] = None,
) -> Tuple[Any, str]:
    url = base_url.rstrip("/") + path
    data = None
    req_headers = headers.copy() if headers else {}
    if payload is not None:
        data = json.dumps(payload).encode()
        req_headers["Content-Type"] = "application/json"

    request = urllib.request.Request(url, data=data, headers=req_headers, method=method)
    with urllib.request.urlopen(request, timeout=30) as response:
        body = response.read().decode("utf-8", errors="replace")
        try:
            parsed = json.loads(body)
        except json.JSONDecodeError as exc:
            raise SmokeCheckError(f"{path} returned invalid JSON: {exc}; body={body[:500]!r}") from exc
        return parsed, body


def expect_mapping(value: Any, path: str) -> Dict[str, Any]:
    if not isinstance(value, dict):
        raise SmokeCheckError(f"{path} returned {type(value).__name__}, expected object")
    return value


def record_code(
    results: Dict[str, Any],
    failures: Dict[str, Any],
    scope: str,
    name: str,
    path: str,
    payload: Any,
    expected_code: Any,
) -> None:
    bucket = results.setdefault(scope, {})
    entry: Dict[str, Any] = {"path": path}

    if isinstance(payload, dict):
        entry["code"] = payload.get("code")
        message = payload.get("message")
        if message is not None:
            entry["message"] = message
        bucket[name] = entry
        if entry["code"] != expected_code:
            failures[f"{scope}.{name}"] = {
                "path": path,
                "expected_code": expected_code,
                "actual_code": entry["code"],
                "message": message,
            }
        return

    entry["error"] = f"invalid response type: {type(payload).__name__}"
    bucket[name] = entry
    failures[f"{scope}.{name}"] = {
        "path": path,
        "error": entry["error"],
    }


def run_admin_smoke(base_url: str, account: str, password: str) -> Tuple[Dict[str, Any], Dict[str, Any]]:
    results: Dict[str, Any] = {}
    failures: Dict[str, Any] = {}

    login_path = "/api/sys/user/login"
    login_payload, _ = request_json(
        base_url,
        login_path,
        method="POST",
        payload={"account": account, "password": password},
    )
    login = expect_mapping(login_payload, login_path)
    record_code(results, failures, "admin", "login", login_path, login, "000000")

    token = expect_mapping(login.get("data"), login_path).get("token")
    if not token:
        raise SmokeCheckError(f"{login_path} succeeded but did not return token")

    checks = {
        "user_info": "/api/sys/user/info",
        "role_list": "/api/sys/role/queryRoleList?current=1&pageSize=10",
        "tenant_list": "/api/sys/tenant/queryTenantList?current=1&pageSize=10",
        "menu_template_list": "/api/sys/menuTemplate/queryMenuTemplateList?current=1&pageSize=10",
        "coupon_list": "/api/sms/coupon/queryCouponList?current=1&pageSize=10",
        "home_advertise_list": "/api/sms/homeAdvertise/queryHomeAdvertiseList?current=1&pageSize=10",
        "home_brand_list": "/api/sms/homeBrand/queryHomeBrandList?current=1&pageSize=10",
        "seckill_activity_list": "/api/sms/seckillActivity/querySeckillActivityList?current=1&pageSize=10",
        "subject_list": "/api/cms/subject/querySubjectList?current=1&pageSize=10",
        "preferred_area_list": "/api/cms/prefrenceArea/queryPreferredAreaList?current=1&pageSize=10",
        "subject_category_list": "/api/cms/subjectCategory/querySubjectCategoryList?current=1&pageSize=10",
        "home_recommend_subject_list": "/api/sms/homeRecommendSubject/queryHomeRecommendSubjectList?current=1&pageSize=10",
    }

    for name, path in checks.items():
        payload, _ = request_json(base_url, path, headers={"Authorization": token})
        record_code(results, failures, "admin", name, path, payload, "000000")

    return results, failures


def run_front_smoke(base_url: str) -> Tuple[Dict[str, Any], Dict[str, Any]]:
    results: Dict[str, Any] = {}
    failures: Dict[str, Any] = {}

    checks = {
        "home_index": "/api/home/index",
        "product_list": "/api/product/queryProductList?productCategoryId=1&current=1&pageSize=10",
        "product_detail": "/api/product/queryProductDetail?id=1",
    }

    for name, path in checks.items():
        payload, _ = request_json(base_url, path)
        record_code(results, failures, "front", name, path, payload, 0)

    return results, failures


def merge_dicts(target: Dict[str, Any], source: Dict[str, Any]) -> None:
    for key, value in source.items():
        if key not in target:
            target[key] = value
        elif isinstance(target[key], dict) and isinstance(value, dict):
            target[key].update(value)
        else:
            target[key] = value


def main() -> int:
    parser = argparse.ArgumentParser(description="Smoke check the Zero Admin remote deployment")
    parser.add_argument("--base-url", default="http://47.107.224.56:8000")
    parser.add_argument("--front-base-url", default="")
    parser.add_argument("--account", default="admin")
    parser.add_argument("--password", default="123456")
    args = parser.parse_args()

    try:
        results: Dict[str, Any] = {}
        failures: Dict[str, Any] = {}

        admin_results, admin_failures = run_admin_smoke(args.base_url, args.account, args.password)
        merge_dicts(results, admin_results)
        failures.update(admin_failures)

        if args.front_base_url:
            front_results, front_failures = run_front_smoke(args.front_base_url)
            merge_dicts(results, front_results)
            failures.update(front_failures)

        if failures:
            print(
                json.dumps(
                    {
                        "results": results,
                        "failures": failures,
                    },
                    ensure_ascii=False,
                ),
                file=sys.stderr,
            )
            return 1

        print(json.dumps(results, ensure_ascii=False))
        return 0
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        print(
            json.dumps(
                {"error": "http_error", "status": exc.code, "body": body},
                ensure_ascii=False,
            ),
            file=sys.stderr,
        )
        return 1
    except Exception as exc:
        print(
            json.dumps({"error": "exception", "message": str(exc)}, ensure_ascii=False),
            file=sys.stderr,
        )
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
