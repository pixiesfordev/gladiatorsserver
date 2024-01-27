$utf8WithoutBom = New-Object System.Text.UTF8Encoding $false

# 更新 Dev_Matchmaker.yaml 文件的版本
$content = [System.IO.File]::ReadAllText('CICD_Matchmaker_Dev\Dev_Matchmaker.yaml', $utf8WithoutBom)
$pattern = 'gladiators-matchmaker:(\d+\.\d+\.)(\d+)'
$match = [regex]::Match($content, $pattern)

if ($match.Success) {
    $versionMajorMinor = $match.Groups[1].Value
    $versionPatch = [int]$match.Groups[2].Value
    $newVersionPatch = $versionPatch + 1
    $newVersion = $versionMajorMinor + $newVersionPatch
    $content = $content -replace $pattern, ('gladiators-matchmaker:' + $newVersion)
    [System.IO.File]::WriteAllText('CICD_Matchmaker_Dev\Dev_Matchmaker.yaml', $content, $utf8WithoutBom)
    Write-Host "Successfully matched and modified the version to: $newVersion"
} else {
    Write-Host 'No matching version found for gladiators-matchmaker in Dev_Matchmaker.yaml'
}

# 更新 Dev_BuildMatchmaker.bat 文件的版本
$content = [System.IO.File]::ReadAllText('CICD_Matchmaker_Dev\Dev_BuildMatchmaker.bat', $utf8WithoutBom)
$pattern = 'gladiators-matchmaker:(\d+\.\d+\.)(\d+)'
$match = [regex]::Match($content, $pattern)

if ($match.Success) {
    $versionMajorMinor = $match.Groups[1].Value
    $versionPatch = [int]$match.Groups[2].Value
    $newVersionPatch = $versionPatch + 1
    $newVersion = $versionMajorMinor + $newVersionPatch
    $content = $content -replace $pattern, ('gladiators-matchmaker:' + $newVersion)
    [System.IO.File]::WriteAllText('CICD_Matchmaker_Dev\Dev_BuildMatchmaker.bat', $content, $utf8WithoutBom)
    Write-Host "Successfully matched and modified the version to: $newVersion"
} else {
    Write-Host 'No matching version found for gladiators-matchmaker in Dev_BuildMatchmaker.bat'
}
