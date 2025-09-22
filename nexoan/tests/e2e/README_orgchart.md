# Organizational Chart Test Suite

This test suite demonstrates a comprehensive organizational chart structure with 3 Ministers and 6 Departments (2 per minister), all containing rich tabular data.

## Test Structure

### OrgChartIngestion Class
The `OrgChartIngestion` class creates and manages organizational chart data with:

- **3 Ministers**: Technology, Health, and Education
- **6 Departments**: 2 departments under each minister
- **Rich Tabular Attributes**: Each entity contains multiple tabular data structures

### Entity Types
- **Ministers**: `{"major": "Organization", "minor": "Minister"}`
- **Departments**: `{"major": "Organization", "minor": "Department"}`
- **All Attributes**: `{"major": "Dataset", "minor": "Tabular"}`

### Ministers
1. **Minister of Technology and Digital Innovation** (`minister-tech-001`)
   - Department of Information and Communication Technology (`dept-ict-001`)
   - Department of Digital Innovation and Research (`dept-innovation-001`)

2. **Minister of Health and Social Services** (`minister-health-001`)
   - Department of Hospital Services (`dept-hospitals-001`)
   - Department of Public Health and Prevention (`dept-public-health-001`)

3. **Minister of Education and Human Development** (`minister-education-001`)
   - Department of School Education (`dept-schools-001`)
   - Department of Higher Education and Research (`dept-higher-ed-001`)

### Type Structure

All entities and attributes follow the specified type system:

- **Entity Types**:
  - Ministers: `Organization/Minister`
  - Departments: `Organization/Department`

- **Attribute Types**:
  - All tabular data: `Dataset/Tabular`

This ensures proper type classification for querying and data processing.

### Tabular Data Structure

#### Minister Attributes
- **Personal Information**: Name, portfolio, contact details, security clearance
- **Performance Metrics**: Budget utilization, policy implementations, approval ratings
- **Budget Allocation**: Operational expenses, capital investments, research grants

#### Department Attributes
- **Department Information**: Name, focus area, establishment date, contact details
- **Staff Information**: Position counts, vacancy data, salary grades
- **Project Portfolio**: Project status, budgets, timelines, progress
- **Budget Breakdown**: Personnel costs, operational expenses, capital expenditure

### Test Methods

1. **`test_orgchart_creation()`**: Creates all entities (3 ministers + 6 departments)
2. **`test_orgchart_query()`**: Basic queries for ministers and departments
3. **`test_tabular_data_validation()`**: Validates tabular data structure and content
4. **`test_department_relationships()`**: Tests department-to-minister relationships
5. **`test_budget_data_analysis()`**: Cross-ministerial budget aggregation
6. **`test_staff_data_aggregation()`**: Department-level staff analysis

## Running the Tests

### Prerequisites
```bash
pip install requests
```

### Environment Variables
Set the following environment variables or use defaults:
- `QUERY_SERVICE_URL`: Default `http://0.0.0.0:8081`
- `UPDATE_SERVICE_URL`: Default `http://0.0.0.0:8080`

### Execute Tests
```bash
python test_orgchart_test.py
```

## Expected Output

The test suite will:
1. Create 9 entities (3 ministers + 6 departments)
2. Validate tabular data structure and content
3. Test hierarchical relationships
4. Perform budget and staff data analysis
5. Display comprehensive test results with metrics

## Data Examples

### Minister Performance Metrics Table
| metric | target | actual | period | status |
|--------|--------|--------|--------|--------|
| budget_utilization | 95% | 92% | Q1-2024 | On Track |
| policy_implementations | 5 | 3 | Q1-2024 | In Progress |
| public_approval_rating | 80% | 78% | Q1-2024 | Good |

### Department Staff Information Table
| position | count | vacant | salary_grade | department |
|----------|-------|--------|--------------|------------|
| Director General | 1 | 0 | SL-1 | ICT Department |
| Deputy Director | 2 | 0 | SL-2 | ICT Department |
| Senior Officer | 15 | 2 | SL-4 | ICT Department |

This test suite demonstrates the power of tabular data in representing complex organizational structures with rich, queryable information.
