#!/usr/bin/python3

import sys
import os

class LinuxTimeStat:
    def __init__(self, path):
        self._filePath = path
        self._realTime = []
        self._userTime = []
        self._sysTime = []
        self._userPlusSys = []
        self._parseFile()

    def getRealTime(self):
        return self._realTime

    def getUserTime(self):
        return self._userTime

    def getSysTime(self):
        return self._sysTime

    def getUserPlusSysTime(self):
        return self._userPlusSys

    def _parseFile(self):
        f = open(self._filePath, "r")
        prevUserTime = 0
        for line in f:
            if (line != "\n"):
                arr = line.split()
                arr2 = arr[1].split("m")
                strMinutes = arr2[0]
                minutes = int(strMinutes)
                strSeconds = arr2[1].replace("s","")
                seconds = float(strSeconds)
                if (arr[0] == "real"):
                    self._realTime.append((minutes*60 + seconds))
                elif (arr[0] == "user"):
                    prevUserTime = minutes*60 + seconds
                    self._userTime.append(prevUserTime)
                elif (arr[0] == "sys"):
                    sysTime = minutes*60 + seconds
                    self._sysTime.append(sysTime)
                    self._userPlusSys.append(prevUserTime + sysTime)
                else:
                    print(line)

        return

arrTimeStat = []
directory = "UserPlusSysTimes"
if not os.path.exists(directory):
    os.makedirs(directory)
for path in sys.argv[1:]:
    print(path)
    arrTimeStat.append(LinuxTimeStat(path))
    newFileName = directory + "/UserSys_" + path.split("_")[-1]
    f = open(newFileName,"w")
    strArr = str(arrTimeStat[-1].getUserPlusSysTime())
    strArr = strArr[1:]
    strArr = strArr[:-1]
    f.write(strArr)
