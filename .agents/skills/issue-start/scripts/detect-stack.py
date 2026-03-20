#!/usr/bin/env python3
# 自动检测项目技术栈，输出 linter 和 test 命令
# Usage: python detect-stack.py [project_root]

import sys
import os


def main():
    project_root = sys.argv[1] if len(sys.argv) > 1 else "."
    os.chdir(project_root)

    linter = ""
    test = ""
    formatter = ""
    stack = ""

    if os.path.isfile("pom.xml") or os.path.isfile("build.gradle"):
        stack = "java"
        if os.path.isfile("pom.xml"):
            linter = "mvn checkstyle:check -q 2>/dev/null || echo 'checkstyle not configured'"
            test = "mvn test -q"
            formatter = "mvn spotless:apply 2>/dev/null || echo 'spotless not configured'"
        else:
            linter = "./gradlew checkstyleMain -q 2>/dev/null || echo 'checkstyle not configured'"
            test = "./gradlew test -q"
            formatter = "./gradlew spotlessApply 2>/dev/null || echo 'spotless not configured'"
    elif os.path.isfile("package.json"):
        stack = "javascript/typescript"
        linter = "npm run lint --silent 2>/dev/null || npx eslint . 2>/dev/null || echo 'no linter configured'"
        test = "npm test -- --coverage 2>/dev/null || echo 'no test configured'"
        formatter = "npx prettier --write . 2>/dev/null || echo 'prettier not configured'"
    elif os.path.isfile("pyproject.toml") or os.path.isfile("requirements.txt"):
        stack = "python"
        linter = "pylint $(git ls-files '*.py' 2>/dev/null) 2>/dev/null || echo 'pylint not available'"
        test = "pytest --maxfail=1 --disable-warnings --cov 2>/dev/null || echo 'pytest not available'"
        formatter = "black . 2>/dev/null || echo 'black not available'"
    elif os.path.isfile("pubspec.yaml"):
        stack = "dart"
        try:
            with open("pubspec.yaml", "r", encoding="utf-8") as f:
                content = f.read().lower()
            if "flutter" in content:
                stack = "flutter"
                linter = "flutter analyze"
                test = "flutter test --coverage"
                formatter = "flutter format ."
            else:
                linter = "dart analyze"
                test = "dart test"
                formatter = "dart format ."
        except (IOError, UnicodeDecodeError):
            linter = "dart analyze"
            test = "dart test"
            formatter = "dart format ."
    elif os.path.isfile("go.mod"):
        stack = "go"
        linter = "golangci-lint run 2>/dev/null || go vet ./..."
        test = "go test -cover ./..."
        formatter = "gofmt -w ."
    elif os.path.isfile("Cargo.toml"):
        stack = "rust"
        linter = "cargo clippy -- -D warnings"
        test = "cargo test"
        formatter = "cargo fmt"

    print(f"STACK={stack or 'unknown'}")
    print(f"LINTER={linter or '<手动设置>'}")
    print(f"TEST={test or '<手动设置>'}")
    print(f"FORMATTER={formatter or '<手动设置>'}")


if __name__ == "__main__":
    main()
