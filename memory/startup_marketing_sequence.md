# Startup-Sequenz für Marketing

Stand: 2026-10-09. Der Nutzer hat die funktionierende Sequenz bestätigt und
wünscht nun fünf Sekunden Dauerlicht ohne Atmen: möglichst satte und helle
Farben entsprechend den Farbtönen aus `12 colored dots.png`.

- Implementierung in `startup_animation.go`, im Produktionsstart aktiviert.
- Trigger: erster Poll mit erreichbarer Geräte-IP pro Gerät und Serverprozess.
  Die dokumentierte API enthält kein dediziertes Boot-Ereignis.
  Ein Geräte-Neustart allein startet die Sequenz bei weiterlaufendem Server
  deshalb nicht erneut; zum Wiederholen den Server neu starten.
- Originalfarben aus den Kreismittelpunkten des Bilds, oben beginnend im
  Uhrzeigersinn:
  `#D21176 #6F157A #6B167A #75127B #870B7A #A6077C`
  `#CD067E #DE2A66 #EF4E4F #F1723E #F4972E #F17741`.
- Für maximale Sättigung wird der gemeinsame RGB-Minimalwert entfernt und
  der größte Kanal auf 255 skaliert. Das war die erste Verstärkung der Bildpalette; die
  ausgegebenen Werte entsprechen bewusst nicht exakt den Bildpixeln:
  `#FF0085 #9000FF #8800FF #9800FF #C000FF #FF00BC`
  `#FF009A #FF0055 #FF0002 #FF4A00 #FF8700 #FF4E00`.
- Weitere Live-Abstimmung: Der Nutzer bestätigt die Verbesserung, findet
  aber die rechten Violetttöne noch zu hell und möchte maximale Farbintensität
  entsprechend dem Markenbild. LEDs 1–4 (oben = 0, im Uhrzeigersinn) erhalten
  weniger Rot, weiterhin Blau 255 und Grün 0, für einen tieferen Violetteindruck.
  Das ist eine vorläufige manuelle Farbkorrektur nach Gerätefeedback, keine
  gemessene Farbkalibrierung. Die bereits maximalen Sättigungs-/Helligkeitswerte
  lassen sich digital nicht weiter steigern.
- Alle zwölf LEDs leuchten konstant. Nach weiterem Live-Feedback sind
  die pinken LEDs oben und links unten (LEDs 0 und 7) auf 70 Prozent
  Helligkeit reduziert, da diese
  Positionen zu dominant wirken; die übrigen bleiben bei 100 Prozent. Nach dem
  ersten akzeptierten Farbframe wird fünf Sekunden gewartet, ohne weitere
  Animationsframes; danach wird der aktuelle normale Ring direkt gepusht.
- Währenddessen liefern zusätzliche Polls die Marketing-Anzeige;
  konkurrierende normale LED-Pushes werden unterdrückt und HTTP-Schreibzugriffe
  serialisiert. Der anschließende Ring wird aus dem aktuellen
  Cache/Buchungszustand neu berechnet.
- Button-/NFC-Events brechen die Sequenz ab und übernehmen die Anzeige.
  Nach Abbruch wird kein zusätzlicher Umgebungsframe mehr gesendet: dieser
  konnte zuvor die Button-Rückmeldung überschreiben. Bei normalem Abschluss
  erfolgt die Wiederherstellung synchron und mit dem Sequenz-Kontext, sodass
  auch ein gleichzeitiger Abbruch einen wartenden Push unterbindet.
  Interaktion widerruft den Sequenz-Kontext auch nach dem sichtbaren Ende,
  falls die Wiederherstellung noch wartet. Dieser
  reine Anzeigepfad löst keinen Commit einer offenen Auswahl aus.
  Die Sequenz führt keine Rooms-Buchungsaktion aus.
- Ohne Gerätecontroller oder bei Loopback/fehlender IP keine Sequenz.
- Die lokale LED-API wird beim ersten Poll bis zu fünf Sekunden lang erneut
  angesprochen, falls sie noch nicht bereit ist. Die sichtbaren fünf Sekunden
  beginnen erst nach dem ersten akzeptierten Farbframe. Scheitert die
  Vorbereitung komplett, wird dies geloggt; kein Neustart auf späteren Polls,
  damit keine verspätete Überraschungsanimation entsteht.
- Konsolenlogs zeigen Warten auf lokale API, Start/Ende sowie einen Abbruch
  durch Gerätebedienung. Der Startlog nennt Dauerlicht und die Helligkeiten 100/70 Prozent.
  Die Nutzermeldung bezeichnet links unten als 8 Uhr; das nachgereichte
  Gerätefoto zeigt dort Pink. Daher wird der pinke Paletteneintrag 7 statt
  des benachbarten roten Eintrags 8 reduziert. Exakte Bild-/Geräteorientierung
  wurde nicht separat kalibriert.
- Windows IPv4-mapped-IPv6-Adressen werden normalisiert; echte IPv6-Ziele
  werden für den HTTP-Rückkanal korrekt geklammert.
- Verifikation: gesättigte Palette, konstante abgestimmte Helligkeiten, reale
  Fünfsekundendauer, nur Start-/Wiederherstellungsframe, einmalige Ausführung,
  Poll-Verhalten, Wiederherstellung und Abbruch automatisiert getestet.
  Die neue Farbwirkung muss noch am echten PuKK beurteilt werden.

- Nutzer meldet nach der Sequenz einen komplett roten Ring und verändertes
  Buttonverhalten. Die letzten Farbänderungen betrafen ausschließlich die
  Startpalette; normale Grün-/Rot-/Orange-/Blau-/Violettwerte blieben gleich.
  Die Abbruch-Rückmeldung wurde als mögliche Störung reproduziert und korrigiert.
  Ein bestätigter freier Raum wird nach der Sequenz grün und ein kurzer Druck
  wählt weiterhin 15 Minuten blau (Regressionstest). Der konkrete Live-Grund
  für den roten Ring ist ohne `/debug/status`/Rooms-Refresh-Logs offen; keine
  künstliche Grün-Anzeige für unbekannte oder belegte Zeitabschnitte einführen.
