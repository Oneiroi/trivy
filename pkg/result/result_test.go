package result

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/aquasecurity/trivy-db/pkg/db"
	dbTypes "github.com/aquasecurity/trivy-db/pkg/types"
	"github.com/aquasecurity/trivy-db/pkg/utils"
	"github.com/aquasecurity/trivy-db/pkg/vulnsrc/vulnerability"
	"github.com/aquasecurity/trivy/pkg/types"
)

func TestClient_FillInfo(t *testing.T) {
	type args struct {
		vulns      []types.DetectedVulnerability
		reportType string
	}
	tests := []struct {
		name                    string
		getSeverity             []db.OperationGetSeverityExpectation
		getVulnerability        []db.OperationGetVulnerabilityExpectation
		args                    args
		expectedVulnerabilities []types.DetectedVulnerability
	}{
		{
			name: "happy path, with only OS vulnerability but no vendor severity, no NVD",
			getVulnerability: []db.OperationGetVulnerabilityExpectation{
				{
					Args: db.OperationGetVulnerabilityArgs{
						VulnerabilityID: "CVE-2019-0001",
					},
					Returns: db.OperationGetVulnerabilityReturns{
						Vulnerability: dbTypes.Vulnerability{
							Title:            "dos",
							Description:      "dos vulnerability",
							Severity:         dbTypes.SeverityMedium.String(),
							References:       []string{"http://example.com"},
							LastModifiedDate: utils.MustTimeParse("2020-01-01T01:01:00Z"),
							PublishedDate:    utils.MustTimeParse("2001-01-01T01:01:00Z"),
						},
					},
				},
			},
			args: args{
				vulns: []types.DetectedVulnerability{
					{VulnerabilityID: "CVE-2019-0001"},
				},
				reportType: vulnerability.RedHat,
			},
			expectedVulnerabilities: []types.DetectedVulnerability{
				{
					VulnerabilityID: "CVE-2019-0001",
					Vulnerability: dbTypes.Vulnerability{
						Title:            "dos",
						Description:      "dos vulnerability",
						Severity:         dbTypes.SeverityMedium.String(),
						References:       []string{"http://example.com"},
						LastModifiedDate: utils.MustTimeParse("2020-01-01T01:01:00Z"),
						PublishedDate:    utils.MustTimeParse("2001-01-01T01:01:00Z"),
					},
					PrimaryURL: "https://avd.aquasec.com/nvd/cve-2019-0001",
				},
			},
		},
		{
			name: "happy path, with only OS vulnerability but no vendor severity, yes NVD",
			getVulnerability: []db.OperationGetVulnerabilityExpectation{
				{
					Args: db.OperationGetVulnerabilityArgs{
						VulnerabilityID: "CVE-2019-0001",
					},
					Returns: db.OperationGetVulnerabilityReturns{
						Vulnerability: dbTypes.Vulnerability{
							Title:       "dos",
							Description: "dos vulnerability",
							Severity:    dbTypes.SeverityUnknown.String(),
							VendorSeverity: dbTypes.VendorSeverity{
								vulnerability.Nvd: dbTypes.SeverityLow,
							},
							References:       []string{"http://example.com"},
							LastModifiedDate: utils.MustTimeParse("2020-01-01T01:01:00Z"),
							PublishedDate:    utils.MustTimeParse("2001-01-01T01:01:00Z"),
						},
					},
				},
			},
			args: args{
				vulns: []types.DetectedVulnerability{
					{VulnerabilityID: "CVE-2019-0001"},
				},
				reportType: vulnerability.Ubuntu,
			},
			expectedVulnerabilities: []types.DetectedVulnerability{
				{
					VulnerabilityID: "CVE-2019-0001",
					Vulnerability: dbTypes.Vulnerability{
						Title:            "dos",
						Description:      "dos vulnerability",
						Severity:         dbTypes.SeverityLow.String(),
						References:       []string{"http://example.com"},
						LastModifiedDate: utils.MustTimeParse("2020-01-01T01:01:00Z"),
						PublishedDate:    utils.MustTimeParse("2001-01-01T01:01:00Z"),
					},
					SeveritySource: vulnerability.Nvd,
					PrimaryURL:     "https://avd.aquasec.com/nvd/cve-2019-0001",
				},
			},
		},
		{
			name: "happy path, with only OS vulnerability but no severity, no vendor severity, no NVD",
			getVulnerability: []db.OperationGetVulnerabilityExpectation{
				{
					Args: db.OperationGetVulnerabilityArgs{
						VulnerabilityID: "CVE-2019-0001",
					},
					Returns: db.OperationGetVulnerabilityReturns{
						Vulnerability: dbTypes.Vulnerability{
							Title:       "dos",
							Description: "dos vulnerability",
							References:  []string{"http://example.com"},
						},
					},
				},
			},
			args: args{
				vulns: []types.DetectedVulnerability{
					{VulnerabilityID: "CVE-2019-0001"},
				},
				reportType: vulnerability.Ubuntu,
			},
			expectedVulnerabilities: []types.DetectedVulnerability{
				{
					VulnerabilityID: "CVE-2019-0001",
					Vulnerability: dbTypes.Vulnerability{
						Title:       "dos",
						Description: "dos vulnerability",
						Severity:    dbTypes.SeverityUnknown.String(),
						References:  []string{"http://example.com"},
					},
					PrimaryURL: "https://avd.aquasec.com/nvd/cve-2019-0001",
				},
			},
		},
		{
			name: "happy path, with only OS vulnerability, yes vendor severity, with both NVD and CVSS info",
			getVulnerability: []db.OperationGetVulnerabilityExpectation{
				{
					Args: db.OperationGetVulnerabilityArgs{
						VulnerabilityID: "CVE-2019-0001",
					},
					Returns: db.OperationGetVulnerabilityReturns{
						Vulnerability: dbTypes.Vulnerability{
							Title:       "dos",
							Description: "dos vulnerability",
							Severity:    dbTypes.SeverityMedium.String(),
							CweIDs:      []string{"CWE-311"},
							VendorSeverity: dbTypes.VendorSeverity{
								vulnerability.RedHat: dbTypes.SeverityLow, // CentOS uses RedHat
							},
							CVSS: map[string]dbTypes.CVSS{
								vulnerability.Nvd: {
									V2Vector: "(AV:N/AC:L/Au:N/C:P/I:P/A:P)",
									V2Score:  4.5,
									V3Vector: "CVSS:3.0/PR:N/UI:N/S:U/C:H/I:H/A:H",
									V3Score:  5.6,
								},
								vulnerability.RedHat: {
									V2Vector: "AV:N/AC:M/Au:N/C:N/I:P/A:N",
									V2Score:  7.8,
									V3Vector: "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
									V3Score:  9.8,
								},
							},
							References: []string{"http://example.com"},
						},
					},
				},
			},
			args: args{
				vulns: []types.DetectedVulnerability{
					{VulnerabilityID: "CVE-2019-0001"},
				},
				reportType: vulnerability.CentOS,
			},
			expectedVulnerabilities: []types.DetectedVulnerability{
				{
					VulnerabilityID: "CVE-2019-0001",
					Vulnerability: dbTypes.Vulnerability{
						Title:       "dos",
						Description: "dos vulnerability",
						Severity:    dbTypes.SeverityLow.String(),
						CweIDs:      []string{"CWE-311"},
						References:  []string{"http://example.com"},
						CVSS: map[string]dbTypes.CVSS{
							vulnerability.Nvd: {
								V2Vector: "(AV:N/AC:L/Au:N/C:P/I:P/A:P)",
								V2Score:  4.5,
								V3Vector: "CVSS:3.0/PR:N/UI:N/S:U/C:H/I:H/A:H",
								V3Score:  5.6,
							},
							vulnerability.RedHat: {
								V2Vector: "AV:N/AC:M/Au:N/C:N/I:P/A:N",
								V2Score:  7.8,
								V3Vector: "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
								V3Score:  9.8,
							},
						},
					},
					SeveritySource: vulnerability.RedHat,
					PrimaryURL:     "https://avd.aquasec.com/nvd/cve-2019-0001",
				},
			},
		},
		{
			name: "happy path light db, with only OS vulnerability, yes vendor severity",
			getVulnerability: []db.OperationGetVulnerabilityExpectation{
				{
					Args: db.OperationGetVulnerabilityArgs{
						VulnerabilityID: "CVE-2019-0001",
					},
					Returns: db.OperationGetVulnerabilityReturns{
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityMedium.String(),
							VendorSeverity: dbTypes.VendorSeverity{
								vulnerability.Ubuntu: dbTypes.SeverityLow,
							},
						},
					},
				},
			},
			args: args{
				vulns: []types.DetectedVulnerability{
					{VulnerabilityID: "CVE-2019-0001"},
				},
				reportType: vulnerability.Ubuntu,
			},
			expectedVulnerabilities: []types.DetectedVulnerability{
				{
					VulnerabilityID: "CVE-2019-0001",
					Vulnerability: dbTypes.Vulnerability{
						Severity: dbTypes.SeverityLow.String(),
					},
					SeveritySource: vulnerability.Ubuntu,
					PrimaryURL:     "https://avd.aquasec.com/nvd/cve-2019-0001",
				},
			},
		},
		{
			name: "happy path light db, with only OS vulnerability, no vendor severity",
			getVulnerability: []db.OperationGetVulnerabilityExpectation{
				{
					Args: db.OperationGetVulnerabilityArgs{
						VulnerabilityID: "CVE-2020-28928",
					},
					Returns: db.OperationGetVulnerabilityReturns{
						Vulnerability: dbTypes.Vulnerability{
							VendorSeverity: dbTypes.VendorSeverity{},
						},
					},
				},
			},
			args: args{
				vulns: []types.DetectedVulnerability{
					{VulnerabilityID: "CVE-2020-28928"},
				},
				reportType: vulnerability.Alpine,
			},
			expectedVulnerabilities: []types.DetectedVulnerability{
				{
					VulnerabilityID: "CVE-2020-28928",
					Vulnerability: dbTypes.Vulnerability{
						Severity: dbTypes.SeverityUnknown.String(),
					},
					SeveritySource: "",
					PrimaryURL:     "https://avd.aquasec.com/nvd/cve-2020-28928",
				},
			},
		},
		{
			name: "happy path, with only library vulnerability",
			getVulnerability: []db.OperationGetVulnerabilityExpectation{
				{
					Args: db.OperationGetVulnerabilityArgs{
						VulnerabilityID: "CVE-2020-0001",
					},
					Returns: db.OperationGetVulnerabilityReturns{
						Vulnerability: dbTypes.Vulnerability{
							Title:       "COVID-19",
							Description: "a nasty virus vulnerability for humans",
							Severity:    dbTypes.SeverityMedium.String(),
							VendorSeverity: dbTypes.VendorSeverity{
								vulnerability.PythonSafetyDB: dbTypes.SeverityCritical,
							},
							References: []string{"https://www.who.int/emergencies/diseases/novel-coronavirus-2019"},
						},
					},
				},
			},
			args: args{
				vulns: []types.DetectedVulnerability{
					{VulnerabilityID: "CVE-2020-0001"},
				},
				reportType: "poetry",
			},
			expectedVulnerabilities: []types.DetectedVulnerability{
				{
					VulnerabilityID: "CVE-2020-0001",
					Vulnerability: dbTypes.Vulnerability{
						Title:       "COVID-19",
						Description: "a nasty virus vulnerability for humans",
						Severity:    dbTypes.SeverityCritical.String(),
						References:  []string{"https://www.who.int/emergencies/diseases/novel-coronavirus-2019"},
					},
					SeveritySource: vulnerability.PythonSafetyDB,
					PrimaryURL:     "https://avd.aquasec.com/nvd/cve-2020-0001",
				},
			},
		},
		{
			name: "GetVulnerability returns an error",
			getVulnerability: []db.OperationGetVulnerabilityExpectation{
				{
					Args: db.OperationGetVulnerabilityArgs{
						VulnerabilityID: "CVE-2019-0004",
					},
					Returns: db.OperationGetVulnerabilityReturns{
						Err: xerrors.New("failed"),
					},
				},
			},
			args: args{
				vulns: []types.DetectedVulnerability{
					{VulnerabilityID: "CVE-2019-0004"},
				},
			},
			expectedVulnerabilities: []types.DetectedVulnerability{
				{VulnerabilityID: "CVE-2019-0004"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDBConfig := new(db.MockOperation)
			mockDBConfig.ApplyGetSeverityExpectations(tt.getSeverity)
			mockDBConfig.ApplyGetVulnerabilityExpectations(tt.getVulnerability)

			c := Client{
				dbc: mockDBConfig,
			}

			c.FillVulnerabilityInfo(tt.args.vulns, tt.args.reportType)
			assert.Equal(t, tt.expectedVulnerabilities, tt.args.vulns, tt.name)
			mockDBConfig.AssertExpectations(t)
		})
	}
}

func TestClient_getPrimaryURL(t *testing.T) {
	type args struct {
		vulnID string
		refs   []string
		source string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "CVE-ID",
			args: args{
				vulnID: "CVE-2014-8484",
				refs:   []string{"http://linux.oracle.com/cve/CVE-2014-8484.html"},
				source: vulnerability.OracleOVAL,
			},
			want: "https://avd.aquasec.com/nvd/cve-2014-8484",
		},
		{
			name: "RUSTSEC",
			args: args{
				vulnID: "RUSTSEC-2018-0017",
				refs:   []string{"https://github.com/rust-lang-deprecated/tempdir/pull/46"},
				source: vulnerability.RustSec,
			},
			want: "https://rustsec.org/advisories/RUSTSEC-2018-0017",
		},
		{
			name: "GHSA",
			args: args{
				vulnID: "GHSA-28fw-88hq-6jmm",
				refs:   []string{},
				source: vulnerability.PhpSecurityAdvisories,
			},
			want: "https://github.com/advisories/GHSA-28fw-88hq-6jmm",
		},
		{
			name: "Debian temp vulnerability",
			args: args{
				vulnID: "TEMP-0841856-B18BAF",
				refs:   []string{},
				source: vulnerability.DebianOVAL,
			},
			want: "https://security-tracker.debian.org/tracker/TEMP-0841856-B18BAF",
		},
		{
			name: "npm",
			args: args{
				vulnID: "NSWG-ECO-516",
				refs: []string{
					"https://hackerone.com/reports/712065",
					"https://github.com/lodash/lodash/pull/4759",
					"https://www.npmjs.com/advisories/1523",
				},
				source: vulnerability.NodejsSecurityWg,
			},
			want: "https://www.npmjs.com/advisories/1523",
		},
		{
			name: "suse",
			args: args{
				vulnID: "openSUSE-SU-2019:2596-1",
				refs: []string{
					"http://lists.opensuse.org/opensuse-security-announce/2019-11/msg00076.html",
					"https://www.suse.com/support/security/rating/",
				},
				source: vulnerability.OpenSuseCVRF,
			},
			want: "http://lists.opensuse.org/opensuse-security-announce/2019-11/msg00076.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			got := c.getPrimaryURL(tt.args.vulnID, tt.args.refs, tt.args.source)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_Filter(t *testing.T) {
	type args struct {
		vulns         []types.DetectedVulnerability
		severities    []dbTypes.Severity
		ignoreUnfixed bool
		ignoreFile    string
		policyFile    string
	}
	tests := []struct {
		name string
		args args
		want []types.DetectedVulnerability
	}{
		{
			name: "happy path",
			args: args{
				vulns: []types.DetectedVulnerability{
					{
						VulnerabilityID:  "CVE-2019-0001",
						PkgName:          "foo",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityLow.String(),
						},
					},
					{
						VulnerabilityID:  "CVE-2019-0002",
						PkgName:          "bar",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityCritical.String(),
						},
					},
					{
						VulnerabilityID:  "CVE-2018-0001",
						PkgName:          "baz",
						InstalledVersion: "1.2.3",
						FixedVersion:     "",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityHigh.String(),
						},
					},
					{
						VulnerabilityID:  "CVE-2018-0001",
						PkgName:          "bar",
						InstalledVersion: "1.2.3",
						FixedVersion:     "",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityCritical.String(),
						},
					},
					{
						VulnerabilityID:  "CVE-2018-0002",
						PkgName:          "bar",
						InstalledVersion: "1.2.3",
						FixedVersion:     "",
						Vulnerability: dbTypes.Vulnerability{
							Severity: "",
						},
					},
				},
				severities:    []dbTypes.Severity{dbTypes.SeverityCritical, dbTypes.SeverityHigh, dbTypes.SeverityUnknown},
				ignoreUnfixed: false,
			},
			want: []types.DetectedVulnerability{
				{
					VulnerabilityID:  "CVE-2018-0001",
					PkgName:          "bar",
					InstalledVersion: "1.2.3",
					FixedVersion:     "",
					Vulnerability: dbTypes.Vulnerability{
						Severity: dbTypes.SeverityCritical.String(),
					},
				},
				{
					VulnerabilityID:  "CVE-2019-0002",
					PkgName:          "bar",
					InstalledVersion: "1.2.3",
					FixedVersion:     "1.2.4",
					Vulnerability: dbTypes.Vulnerability{
						Severity: dbTypes.SeverityCritical.String(),
					},
				},
				{
					VulnerabilityID:  "CVE-2018-0002",
					PkgName:          "bar",
					InstalledVersion: "1.2.3",
					FixedVersion:     "",
					Vulnerability: dbTypes.Vulnerability{
						Severity: dbTypes.SeverityUnknown.String(),
					},
				},
				{
					VulnerabilityID:  "CVE-2018-0001",
					PkgName:          "baz",
					InstalledVersion: "1.2.3",
					FixedVersion:     "",
					Vulnerability: dbTypes.Vulnerability{
						Severity: dbTypes.SeverityHigh.String(),
					},
				},
			},
		},
		{
			name: "happy path with ignore-unfixed",
			args: args{
				vulns: []types.DetectedVulnerability{
					{
						VulnerabilityID:  "CVE-2019-0001",
						PkgName:          "foo",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityLow.String(),
						},
					},
					{
						VulnerabilityID:  "CVE-2018-0002",
						PkgName:          "bar",
						InstalledVersion: "1.2.3",
						FixedVersion:     "",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityHigh.String(),
						},
					},
				},
				severities:    []dbTypes.Severity{dbTypes.SeverityHigh},
				ignoreUnfixed: true,
			},
		},
		{
			name: "happy path with ignore-file",
			args: args{
				vulns: []types.DetectedVulnerability{
					{
						// this vulnerability is evaluate
						VulnerabilityID:  "CVE-2019-0001",
						PkgName:          "foo",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityLow.String(),
						},
					},
					{
						// this vulnerability is evaluate
						VulnerabilityID:  "CVE-2019-0002",
						PkgName:          "foo",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityLow.String(),
						},
					},
					{
						VulnerabilityID:  "CVE-2019-0003",
						PkgName:          "foo",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityLow.String(),
						},
					},
				},
				severities:    []dbTypes.Severity{dbTypes.SeverityLow},
				ignoreUnfixed: false,
				ignoreFile:    "testdata/.trivyignore",
			},
			want: []types.DetectedVulnerability{
				{
					VulnerabilityID:  "CVE-2019-0003",
					PkgName:          "foo",
					InstalledVersion: "1.2.3",
					FixedVersion:     "1.2.4",
					Vulnerability: dbTypes.Vulnerability{
						Severity: dbTypes.SeverityLow.String(),
					},
				},
			},
		},
		{
			name: "happy path with a policy file",
			args: args{
				vulns: []types.DetectedVulnerability{
					{
						VulnerabilityID:  "CVE-2019-0001",
						PkgName:          "foo",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityLow.String(),
						},
					},
					{
						// this vulnerability is evaluate
						VulnerabilityID:  "CVE-2019-0002",
						PkgName:          "foo",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityLow.String(),
						},
					},
					{
						// this vulnerability is evaluate
						VulnerabilityID:  "CVE-2019-0003",
						PkgName:          "foo",
						InstalledVersion: "1.2.3",
						FixedVersion:     "1.2.4",
						Vulnerability: dbTypes.Vulnerability{
							Severity: dbTypes.SeverityLow.String(),
						},
					},
				},
				severities:    []dbTypes.Severity{dbTypes.SeverityLow},
				ignoreUnfixed: false,
				policyFile:    "./testdata/test.rego",
			},
			want: []types.DetectedVulnerability{
				{
					VulnerabilityID:  "CVE-2019-0001",
					PkgName:          "foo",
					InstalledVersion: "1.2.3",
					FixedVersion:     "1.2.4",
					Vulnerability: dbTypes.Vulnerability{
						Severity: dbTypes.SeverityLow.String(),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			got, err := c.Filter(context.Background(), tt.args.vulns, tt.args.severities,
				tt.args.ignoreUnfixed, tt.args.ignoreFile, tt.args.policyFile)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}