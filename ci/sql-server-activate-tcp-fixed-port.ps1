# Configure SQL server so it accepts TCP connections on the AppVeyor
# CI platfrom.
#
# Pass the script the instanceName and tcpPort
#
# See
# http://www.appveyor.com/docs/services-databases#enabling-tcp-ip-named-pipes-and-setting-instance-alias
# https://gist.githubusercontent.com/FeodorFitsner/d971c5a98782d211640d/raw/sql-server-ip-and-alias.ps1
# http://geekswithblogs.net/TedStatham/archive/2014/06/13/setting-the-ports-for-a-named-sql-server-instance-using.aspx

Param(
  [string]$instanceName,
  [string]$tcpPort
)
[reflection.assembly]::LoadWithPartialName("Microsoft.SqlServer.Smo") | Out-Null
[reflection.assembly]::LoadWithPartialName("Microsoft.SqlServer.SqlWmiManagement") | Out-Null

$serverName = $env:COMPUTERNAME
$smo = 'Microsoft.SqlServer.Management.Smo.'
$wmi = new-object ($smo + 'Wmi.ManagedComputer')

# Enable TCP/IP
$uri = "ManagedComputer[@Name='$serverName']/ServerInstance[@Name='$instanceName']/ServerProtocol[@Name='Tcp']"
$Tcp = $wmi.GetSmoObject($uri)
$Tcp.IsEnabled = $true
foreach ($ipAddress in $Tcp.IPAddresses)
{
    $ipAddress.IPAddressProperties["TcpDynamicPorts"].Value = ""
    $ipAddress.IPAddressProperties["TcpPort"].Value = $tcpPort
}
$Tcp.alter()

# Service needs to be restarted
# Restart service
Restart-Service "MSSQL`$$instanceName"
