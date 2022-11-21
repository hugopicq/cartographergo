param(
    [Parameter(Mandatory = $true)][string]$nmapfile,
    [Parameter(Mandatory = $false)][string]$cartofile,
    [Parameter(Mandatory = $false)][string]$IPHeader = "IP",
    [Parameter(Mandatory = $false)][string]$PortHeader = "OpenPorts",
    [Parameter(Mandatory = $true)][string]$outputfile,
    [Parameter(Mandatory = $false)][switch]$help
)

Function Write-To-Csv {
    param($array, $outfile, $headers, $no_header_line)

    $header_line = ""
    if($headers){
        $properties = $headers
    }
    else {
        #What if no object ?
        $properties = $array[0].PSObject.properties | select -ExpandProperty Name
    }

    foreach($property in $properties){
        if($property -ne $properties[0]){
            $header_line += ";"
        }
        $header_line += $property
    }

    if(!$no_header_line){
        if($encoding){
            $header_line | Out-File -FilePath $outfile -Encoding ASCII
        }
        else {
            $header_line | Out-File -FilePath $outfile -Encoding ASCII
        }
    }

    foreach($object in $array){
        $objectline = ""
        foreach($property in $properties){
            if(!($object."$property") -or $object."$property".ToString() -eq ""){
                $objectline += ";"
            }
            else {
                if($object."$property" -match ';'){
                    $object."$property" = $object."$property" -replace '"','""'
                    $objectline += ('"' + $object."$property" + '";')
                }
                else {
                    $objectline += (([String]$object."$property") + ';')
                }
            }
        }
        $objectline = $objectline.Substring(0, $objectline.Length - 1)
        if($encoding){
            $objectline | Out-File -FilePath $outfile -Encoding ASCII -Append
        }
        else {
            $objectline | Out-File -FilePath $outfile -Encoding ASCII -Append
        }
    }

    if(Test-Path -Path $outfile){
        $MyRawString = Get-Content -Raw $outfile
        $Utf8NoBomEncoding = New-Object System.Text.UTF8Encoding $False
        [System.IO.File]::WriteAllLines($outfile, $MyRawString.Substring(0, $MyRawString.Length - 2), $Utf8NoBomEncoding)
    }
}

Function Log-Entry {
    param($message)
    $date = Get-Date
    Write-Host "[$($date.tostring('HH:mm:ss'))] $message"
}

if($help) {
    Write-Host "Usage: .\NmapToCarto.ps1 -nmapfile <path> -outputfile <path> [-cartofile <path>] [-IPHeader `"IPAddress`"] [-PortHeader `"Ports`"]"
    exit 0
}

if(Test-Path -Path $outputfile){
    $override = Read-Host "An output file was found, do you want to override this file ? (Pressing n will stop the script, otherwise it will be overwritten) (y/n)"
    if($override -eq "y" -and (Test-Path -Path $outputfile)){
        Remove-Item -Path $outputfile
    }
    else {
        Log-Entry "Stopping script"
        exit 0
    }
}

$existingData = @()
$existingHeaders = @("IP", "OpenPorts")

if($cartofile -ne $null -and $cartofile -ne "" -and (Test-Path -Path $cartofile)) {
    Log-Entry "Parsing existing cartographer file..."
    try {
        $data = Get-Content $cartofile
    }
    catch {
        Write-Error "Could not read $cartofile : $($_.Exception.Message)"
        exit 1
    }
    $first = $true
    foreach($datum in $data) {
        if($first) {
            $existingHeaders = $datum -Split ";"
            $first = $false
            continue
        }
        $server = [PSCustomObject]@{}
        $fieldsValue = $datum -Split ";"
        for($i = 0; $i -lt $existingHeaders.Length; $i++) {
            $server | Add-Member -NotePropertyName $existingHeaders[$i] -NotePropertyValue $fieldsValue[$i]
        }
        $existingData += $server
    }
}

Log-Entry "Parsing Nmap file..."

$nmapData = @()

try {
    [Xml]$nmapresults = Get-Content $nmapfile
}
catch {
    Write-Error "Could not read $nmapfile : $($_.Exception.Message)"
    exit 1
}

foreach($hote in $nmapresults.nmaprun.host){
	try {
		$ipaddress = ($hote.address | Where-Object { $_.addrtype -eq "ipv4" }).addr
	}
	catch {
		Log-Entry "!!! No result for a host"
		continue
	}

	$ports = "|" + (($hote.ports.port | Where-Object { $_.state.state -eq "open" }).portid -join "|") + "|"
	
    $nmapObject = [PSCustomObject]@{}
    $nmapObject | Add-Member -NotePropertyName $IPHeader -NotePropertyValue $ipaddress
    $nmapObject | Add-Member -NotePropertyName $PortHeader -NotePropertyValue $ports
    $nmapData += $nmapObject
}

Log-Entry "Sorting Data..."
$existingData = $existingData | Sort-Object -Property $IPHeader
$nmapData = $nmapData | Sort-Object -Property $IPHeader

Log-Entry "Processing all data..."
$i_carto = 0
$dataToAdd = @()

foreach($nmapObject in $nmapData){
    while($nmapObject.$IPHeader -gt $existingData[$i_carto].$IPHeader -and $i_carto -lt $existingData.Length) {
        $i_carto += 1
    }
    #Here we either have Sames IPs, or Carto.IP > Nmap.IP (meaning it's a new IP), or we saw all carto data
    if($nmapObject.$IPHeader -eq $existingData[$i_carto].$IPHeader) {
        #The IP already exists and we consider that cartographer is right
        continue
    } else {
        $dataToAdd += $nmapObject
    }
}

$existingData += $dataToAdd

Log-Entry "Writting results..."

#We rewrite output
Write-To-Csv -array $existingData -outfile $outputfile -headers $existingHeaders

Log-Entry "Done !"