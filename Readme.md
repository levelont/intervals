# Interval merger

Nimmt eine Liste von Intervallen und fügt alle zusammen, die sich überlappen.
Das Ergebnis beeinhaltet sowohl die gemergte Intervale als auch alle nicht überlappenden Intervalle, welche unberührt bleiben.

Beispiel:

```
Input: [25,30] [2,19] [14, 23] [4,8]
Output: [2,23] [25,30]
```

## Programm Ausführung

**Voraussetzung**: Golang 1.20.x muss auf dem System installiert werden.

### String Mode

```
> go run . "[25,30] [2,19] [14, 23] [4,8]"
[2,23] [25,30]
```

### File Mode

```console
> go run . -f data/coding_challenge.txt
result written to file "result.txt"
> cat result.txt
[2,23] [25,30]
```

## Annahmen

- Die Intervalliste-Eingabe ist ein String, und wird zwischen Einführungszeichen als ein Parameter eingegeben.
- Ein Interval folgt die Mathematische Notation `[linker-Randwert,rechter-Randwert]`.
- Intervale sind durch ein, mehrere, oder kein Leerzeichen getrennt.
- Es ist möglich, wie im Bespiel, dass die Intervale selbst Leerzeichen enthalten - siehe Wert `[14, 23]`.
- Intervale sind [abgeschlossen](https://de.wikipedia.org/wiki/Intervall_(Mathematik)#Abgeschlossenes_Intervall).
- Die Randwerte sind naturliche Zahlen zwischen `-9223372036854775808` und `9223372036854775807` - math.MinInt64 und math.MaxInt64.
- Ein Intervall in der Liste braucht maximal 64 Zeichen, sprich 64 Bytes in UTF-8 Encodierung, inkl. beide Randwerte, beide eckige Klammern, die Trennkomma und evtl. vorkommende Leerzeichen.

## Implementierungsdetails

### Kleine Eingaben: String Bearbeitung

Meine Lösing basiert darauf, dass wenn eine Intervalliste nach Rechte-Randwerte aufsteigend sortiert wird, liegen überlappende Intervalle direkt neben einander.

Nach der Sortierung kann die Intervallliste in einem einzigen durchlauf volständig gemerged werden: Konsekutive überlappende Intervalle werden gemerged, nicht überlappende Intervalle werden beibehalten.

### Große Eingaben: File Bearbeitung

Da die Aufgabe die Robistheit-Frage mit Hinblick auf sehr große Eingaben stellt, habe ich mich gedanken über den Fall gemacht, dass die gesammte Intervallliste im Speicher nicht passt.
Mit dem File Mode schlage ich einen Map-Reduce orientierten Ansatz vor:

- Segmente der Liste werden mit ihre "Breite" aufgeschlüsselt: ein representatives Interval `[x,y]`, wo `x` das *kleinste* Linksrandwert des gesamt Segmentes ist, `y` das *großte* Rechtsrandwert.
- Wenn sich die Breiten zwei Segmenten überlappen, mussen die Intervalllisten beiden Segmenten überlappungen haben

Nach und nach können so, auf Basis der `merge()` Funktion, die Segmente zu einem einzigen Ergebnis zusammengefügt werden.

Mit der Umgebungsvariable `FILE_CHUNK_SIZE_MB` kann die Segmentengröße in MB spezifiziert werden. Eine Große von 10MB wird per Default benutzt.

Als Beispiel, so kann ein großes File in Segmenten von 100MB abgearbeitet werden:

```console
> FILE_CHUNK_SIZE_MB=100 go run . -f large_file
```

#### Testing

Das `cmd/generate_test_data.go` kann dafür benutzt werden, Testdata zu generieren. Das Tool gibt auch zurück die Information, wie viele nicht überlappende Intervalle im File enthalten sind:

```console
> go run cmd/generate_test_data.go
Output written to "test_data.txt"
Number of non-overlapping intervals: 49997743
```

Mit den default Parametern, erstellt das Tool ein File etwa 1.8GB groß.

### Bearbeitungszeit

Für die Basis Funktionalität der `merge()` Funktion habe ich 2-3 Stunden benötigt. Insgesammt, inkl. meines Map-Reduce Ansatzes habe ich etwa 16-17 Stunden investiert.

## Wie ist die Laufzeit Ihres Programms ? 

### Merge

Sei `n` die Anzahl von Eingabeintervale. Folgende Faktoren definieren die Laufzeit des `merge()` Algorithmus:

- Go's sort.Slice() Funktion hat eine Laufzeitkomplezität von `O(n * log n)`.
  - siehe [die `pdqsort_func` Dokumentation](https://github.com/golang/go/blob/83c4e533bcf71d86437a5aa9ffc9b5373208628c/src/sort/zsortfunc.go#L61C13-L61C13) und [das Pattern-Defeating Quicksort Paper](https://arxiv.org/pdf/2106.05123.pdf).
- Die innere for-Schleife läuft die Eingabeliste einmal durch, was eine Laufzeitkomplezität von `O(n)` entspricht.

Von daher hat `merge()` ein asymptotisches Laufzeitverhalten von `O(n * log n + n)`.

### Parse

Sei `n` die Anzahl von Eingabeintervale. Da `parse()` der Eingabestring auf Intervalebene bearbeitet, hat es auch eine Laufzeitkomplexität von `O(n)`.

### ProcessFile

Die Komplezität von `processFile()` hängt von der Komplexität der von ihr aufgerufenen Funktionen:

- `splitFile()`: Es werden `s` Segmente des Eingabefiles bearbeitet. Für jedes Segment wird die "Breite" berechnet, welche `n/s` Iterationen braucht. Von daher: `O(n)`
- `slices.Sort()` für `s` elemente: `O(s * log s)`. 
- Die Haupt for-Schleife iteriert `s-1` Male, und ruft, im Worst-Case-Scenario, jedes Mal `mergeIntervalsFromFiles()` auf.
- `mergeIntervalsFromFiles()` bearbeitet zwei Files und ruft `parse()` für jedes File. Jedes File beeinhaltet `n/s` Intervalle, was eine Parsing-Komplexität von `O(2n/s)` ergibt. Dazu kommt die Komplexität vom `merge()` Aufruf für den entsprechenden Subset von Intervalle: `O(2n/s * log 2n/s + 2n/s)`.

So ergibt sich eine gesammte Laufzeitkomplexität von 
`O(s + s * log s + (s-1)(2n/s + 2n/s * log 2n/s + 2n/s))`


## Wie kann die Robustheit sichergestellt werden, vor allem auch mit Hinblick auf sehr große Eingaben ?

Mit dem vorgeschlagenen Map-Reduce Ansatz - siehe [Große Eingaben.](###-Große-Eingaben-File-Beabeitung)

## Wie verhält sich der Speicherverbrauch Ihres Programms ?

### Merge

Sei `n` die Anzahl von eingegebenen Intervale. Die eingegebene Intervallliste wird in einem Slice für die Bearbeitung gespeichert.
Da `merge()` arbeitet direkt auf dem Slice, verbraucht es kein zusätzlicher Speicher um Intervale zu mergen.

Somit hat die Funktion ein Speicherverbrauch von `O(n)`.

### Parse

`parse()` berechnet aus dem Eingabestring wie viele Intervale es enthält und stellt ausreichender Speicherplatz dafür bereit. Von daher hat es ein Speicherverbrauch von `O(n)`.

### ProcessFile

Der Speicherverbrauch von `processFile()` hängt von dem Konsum der von ihr aufgerufenen Funktionen:

- `splitFile()`: Sei `f` die Große vom EingabeFile und `s` die über `MAX_CHUNK_FILE_SIZE` gegebenen Segmentgröße. Es wird zweckst Scanning des Eingabefiles Ein Buffer von Große `s` erstellt. Darüberhinaus wird ein dynamisches Array erstellt um das Index aufzubauen, der maximal `f/s` Elemente haben wird. So ergibt sich ein Speicherverbrauch von `O(s+f/s)`
- Die Haupt for-Schleife iteriert `(f/s)-1` Male, und ruft, im Worst-Case-Scenario, jedes Mal `mergeIntervalsFromFiles()` auf.
- `mergeIntervalsFromFiles()` bearbeitet zwei Files und behält ihren Inhalt im Speicher, was `O(2f/s)` Speicher verbraucht. Jeder Inhalt wird geparsed, was zusätzlicher `O(2f/s)` Speicher verbraucht. . Dazu kommt der Verbrauch vom `merge()` Aufruf für den entsprechenden Subset von Intervalle: `O(2f/s * log 2f/s + 2f/s)`.

So ergibt sich eine gesammte Laufzeitkomplexität von 
`O(s + (f/s) + (f/s-1)(2f/s + 2f/s + 2f/s * log 2f/s + 2f/s))`

## Danke!

Vielen Dank für die Gelegenheit, mich bei Euch zu bewerben und meine Lösung für diese Coding Task vorzustellen. Ich bedanke mich ebenfalls im Voraus für Eure Bemühungen, diese zu überprüfen.

Die Aufgabe war Interessant, und es hat Spaß gemacht daran zu arbeiten. Ich freue mich sehr auf Euren Feedback!

Mit den besten Wünschen,

~Luis