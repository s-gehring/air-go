#!/bin/bash
# T100: Verify GraphQL schema matches gql-specs/ documentation for all 5 search queries

echo "Verifying search query schemas..."
echo ""

SCHEMA_FILE="schema.graphqls"
FAILED=0

# Function to verify query signature
verify_query() {
    local query_name=$1
    local filter_type=$2
    local sorter_type=$3
    local output_type=$4

    echo "Checking $query_name..."

    # Extract query signature
    signature=$(grep -A 7 "${query_name}(" $SCHEMA_FILE)

    if [ -z "$signature" ]; then
        echo "  ❌ FAIL: $query_name not found in schema"
        FAILED=1
        return
    fi

    # Check required parameters
    if echo "$signature" | grep -q "where: ${filter_type}"; then
        echo "  ✓ where parameter correct ($filter_type)"
    else
        echo "  ❌ FAIL: where parameter incorrect or missing (expected $filter_type)"
        FAILED=1
    fi

    if echo "$signature" | grep -q "order: \[${sorter_type}!\]"; then
        echo "  ✓ order parameter correct ($sorter_type)"
    else
        echo "  ❌ FAIL: order parameter incorrect or missing (expected [$sorter_type!])"
        FAILED=1
    fi

    if echo "$signature" | grep -q "first: Long"; then
        echo "  ✓ first parameter correct"
    else
        echo "  ❌ FAIL: first parameter missing"
        FAILED=1
    fi

    if echo "$signature" | grep -q "after: String"; then
        echo "  ✓ after parameter correct"
    else
        echo "  ❌ FAIL: after parameter missing"
        FAILED=1
    fi

    if echo "$signature" | grep -q "last: Long"; then
        echo "  ✓ last parameter correct"
    else
        echo "  ❌ FAIL: last parameter missing"
        FAILED=1
    fi

    if echo "$signature" | grep -q "before: String"; then
        echo "  ✓ before parameter correct"
    else
        echo "  ❌ FAIL: before parameter missing"
        FAILED=1
    fi

    if echo "$signature" | grep -q "${output_type}!"; then
        echo "  ✓ return type correct ($output_type!)"
    else
        echo "  ❌ FAIL: return type incorrect (expected $output_type!)"
        FAILED=1
    fi

    echo ""
}

# Verify all 5 search queries with correct type names from schema
verify_query "customerSearch" "CustomerQueryFilterInput" "CustomerQuerySorterInput" "QueryOutputOfCustomer"
verify_query "employeeSearch" "EmployeeQueryFilterInput" "EmployeeQuerySorterInput" "QueryOutputOfEmployee"
verify_query "teamSearch" "TeamQueryFilterInput" "TeamQuerySorterInput" "QueryOutputOfTeamQueryOutput"
verify_query "executionPlanSearch" "ExecutionPlanQueryFilterInput" "ExecutionPlanQuerySorterInput" "QueryOutputOfExecutionPlan"
verify_query "referencePortfolioSearch" "ReferencePortfolioQueryFilterInput" "ReferencePortfolioQuerySorterInput" "QueryOutputOfReferencePortfolioOutput"

# Summary
echo "========================================="
if [ $FAILED -eq 0 ]; then
    echo "✅ ALL SEARCH QUERIES VERIFIED SUCCESSFULLY"
    echo "All 5 entity search queries match the expected schema"
    exit 0
else
    echo "❌ VERIFICATION FAILED"
    echo "Some search queries do not match the expected schema"
    exit 1
fi
