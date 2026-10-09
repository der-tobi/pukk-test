# LinkedIn-Serie: PuKK Workspace x 3V Rooms PoC

Stand: 2026-09-09\
Ziel: mit einer kleinen Serie herausfinden, welche Interaktion am Meetingraum für potenzielle Kunden wirklich nützlich wäre: Check-in, Ad-hoc-Buchen/Verlängern oder Terminieren.

## Entscheidung: Serie statt Gesamtvideo

Die vier POC-Funktionen lösen unterschiedliche Probleme. Ein einziges Gesamtvideo würde zwar die Technik zeigen, aber kaum verraten, **welche** Interaktion bei potenziellen Kunden auf Resonanz stößt. Die Serie trennt diese Hypothesen sauber.

| Video | Eine Hypothese | Frage am Ende |
| --- | --- | --- |
| 1: Check-in | Ein sichtbarer NFC-Check-in hilft gegen ungenutzte, aber reservierte Räume. | „Habt ihr ein No-show-Problem – und würde das helfen?“ |
| 2: Buchen & Verlängern | Direkte Auswahl freier Zeit am Raum ist schneller oder verständlicher als der heutige Weg. | „Wäre das bei euch hilfreich – und wer dürfte es tun?“ |
| 3: Terminieren | Früher freigegebene Räume sollten sofort sichtbar und nutzbar werden. | „Passiert das bei euch oft genug, dass sich diese Interaktion lohnt?“ |
| Optional 4: Weitere Orte | Erst wenn es für Meetingräume Resonanz gibt, ist ein zweiter Einsatzort relevant. | „Wo müsste so eine Interaktion als Nächstes hin?“ |

**Empfehlung:** Alle Clips in einem Drehblock aufnehmen, aber mit vier bis sieben Tagen Abstand veröffentlichen. Jede Folge muss ohne die vorherigen verständlich sein. Nicht „Teil 1 von 3“ in den Hook schreiben; klein reicht „PuKK PoC: Check-in“, „PuKK PoC: Buchen“ und „PuKK PoC: Terminieren“.

Die 35–45-Sekunden-Gesamtdemo bleibt als Reserve: für das Featured-Profil, als Antwort auf „Wie funktioniert das zusammen?“ oder für eine direkte Nachricht an interessierte Personen. Sie ist nicht der erste Validierungspost.

## Klare POC-Grenze

Die Demo ist ein **eigener Integrations-POC** mit PuKK Workspace, 3V Rooms und einer lokal laufenden Go-Anwendung. Farben, Zeitlogik und Interaktionen sind Entscheidungen dieses PoCs. Sie sind keine offizielle Standardintegration, Roadmap, Preis-, Support-, Datenschutz- oder Verfügbarkeitsaussage von proDVX oder 3V Rooms.

In Sekunde 0 und im Schlussbild jedes Videos:

> **EIGENER POC – nicht produktreif**

Sichere Sprache:

- „Ich teste einen eigenen POC mit PuKK Workspace und 3V Rooms.“
- „Im PoC steht Orange für Check-in offen, Rot für eingecheckt bzw. belegt, Blau für eine vorläufige Auswahl und Grün für frei.“
- „Der 3-Sekunden-Langdruck wird vom Gerät nach dem Loslassen als Ereignis gemeldet.“
- „Ich möchte herausfinden, ob diese Interaktion in echten Arbeitsabläufen Reibung nimmt.“

Nicht sagen:

- „die neue proDVX-3V-Rooms-Integration“;
- „produktionsreif“, „Echtzeit“ oder „spart nachweislich Kosten“;
- Aussagen zu Sicherheit, Compliance oder Kundennutzen ohne Beleg;
- Namen, echte Raumdaten oder produktive Kalenderinformationen.

## POC-Funktionen und visuelle Beweise

| Funktion | Sichtbarer Beweis | Ehrliche Grenze |
| --- | --- | --- |
| Status der nächsten Stunde | 12 LEDs zeigen die rollierende nächste Stunde: Grün frei, Orange Check-in offen, Rot aktuell belegt, Violett später belegt. | Farb- und Zeitlogik sind eigene POC-Logik. |
| NFC-Check-in | Orange wechselt nach NFC-Scan animiert zu Rot. | Nur zeigen, wenn der Ablauf mit einer Testbuchung live läuft. |
| Verlängern | Bei einer laufenden Buchung werden freie Anschlusssegmente nach Tastendruck Blau und nach der Bestätigung Rot. | Der PoC übergibt nach fünf Sekunden ohne weiteren Tastendruck; er verlängert nur bis zum nächsten Termin oder zur sichtbaren Stundengrenze. |
| Ad-hoc-Buchung | Bei grünem Ring setzt dieselbe Taste freie Zeit ab jetzt Blau und anschließend auf belegt. | Als kurzer Kontrast im zweiten Video, nicht als eigene vierte Hauptfolge. |
| Terminieren | Nach 3-Sekunden-Langdruck und Loslassen läuft die aktuelle Belegung Blau zurück und nach erfolgreichem Checkout/Release Grün. | Es gibt keine Live-Animation während des Gedrückthaltens, weil das Gerät erst nach dem Loslassen das klassifizierte Ereignis liefert. |

proDVX beschreibt den PuKK Workspace als kompaktes Interaktionsgerät mit 12 einzeln steuerbaren RGB-LEDs, Taste, NFC, WLAN, Bluetooth und REST-API. Die Hardware liefert damit die Bausteine; die gezeigte 3V-Rooms-Logik ist der eigene POC. [Produktseite](https://www.prodvx.com/de/produkte/7031200-pukk-workspace) · [Getting started](https://www.prodvx.com/de/getting-started-with-the-pukk-workspace) · [Remote-Server-API](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_Event_Specification.html)

## Gemeinsamer Drehplan

### Format und Bildsprache

- Hochformat 9:16, 1080 × 1920, 30 fps.
- Pro Folge 20–30 Sekunden, bis auf Video 2 maximal 32 Sekunden.
- Jede Folge startet mit dem Farbwechsel, nicht mit einem Logo oder einer Begrüßung.
- Untertitel einbrennen; der Clip muss ohne Ton verständlich sein.
- Eine ruhige Voice-over-Spur genügt. Das echte Klick- oder Scan-Geräusch darf im ersten Moment stehen bleiben.
- Der Ring ist der Protagonist: Nahaufnahme, echte Hand als Maßstab, neutraler Raumkontext.

LinkedIn empfiehlt, früh zu hooken, jeweils eine klare Idee zu behandeln, vertikal zu filmen, Untertitel zu verwenden und Inhalte im Voraus gebündelt zu produzieren. [LinkedIn: Video im Feed](https://www.linkedin.com/business/talent/blog/talent-acquisition/how-create-videos-for-linkedin-feed), [LinkedIn: B2B-Video-Tipps](https://www.linkedin.com/business/marketing/blog/content-marketing/13-top-tips-for-compelling-b2b-video-content-on-linkedin).

### Kamera und Licht

Das **Handy** ist die Hauptkamera: hochkant auf einem Stativ, Rückkamera auf 1×, Fokus und Belichtung auf den Ring sperren. Die Sony a6500 macht zwei bis vier kurze B-Roll-Nahaufnahmen: NFC-Scan, Finger auf Taste und LED-Ring. Bei der Sony nicht mit f/1.4 filmen; f/2.8 bis f/4 hält Hand und LEDs scharf. Die GoPro bleibt für diese Serie im Schrank. Ein 15-Minuten-Countdown erklärt die Interaktion nicht und bremst den Feed.

1. PuKK auf einen matten Untergrund nahe eines Fensters stellen; Licht kommt seitlich oder leicht von vorne.
2. Neutralen Hintergrund wählen; keine lesbaren Kalender, Namen, Firmenlogos oder echten Buchungsdaten.
3. Testzustände vorbereiten: Grün/frei, Orange/Check-in offen, Rot/eingecheckt, Blau/Verlängerung, Blau/Ad-hoc-Buchung, Grün nach Terminierung und optional Violett/späterer Termin.
4. Jede Aktion dreimal aufnehmen: zwei Sekunden Zustand davor, Aktion, zwei Sekunden Zustand danach.
5. Pro Folge eine Gesamtaufnahme und zwei Nahaufnahmen aufnehmen. So bleiben alle drei Clips visuell verwandt, aber nicht identisch.

## Video 1: Check-in

**Hook:** „Ein Raum kann gebucht sein und trotzdem leer bleiben.“

**Hypothese:** Ein sichtbarer Check-in direkt am Raum löst ein echtes No-show- oder Unsicherheitsproblem.

| Zeit | Bild und Aktion | Sprechertext / Untertitel |
| --- | --- | --- |
| 0:00–0:03 | Extreme Nahaufnahme: Orange wechselt direkt durch NFC-Scan zu Rot. | Oben: „EIGENER POC – nicht produktreif“<br>„Gebucht heißt nicht automatisch: genutzt.“ |
| 0:03–0:08 | Halbnahe Aufnahme am Test-Meetingraum; orange und grüne Ringsegmente sichtbar. | „Ich teste einen physischen Check-in direkt am Raum.“ |
| 0:08–0:15 | NFC-Nahaufnahme noch einmal in Echtzeit. | „Orange: Termin braucht Check-in. Ein Tap bestätigt ihn im POC.“ |
| 0:15–0:22 | Ruhige rote Ringaufnahme. | „Danach ist die aktuelle Belegung sichtbar bestätigt.“ |
| 0:22–0:28 | Grüner oder rot/grüner Ring als Schlussbild. | „Habt ihr ein No-show-Problem – und würde so ein Check-in helfen?“ |

**Voice-over:**

> Ein Raum kann gebucht sein und trotzdem leer bleiben. Ich teste gerade einen eigenen POC mit PuKK Workspace und 3V Rooms. Orange heißt: Der Termin braucht noch Check-in. Ein NFC-Tap bestätigt ihn im POC und macht die aktuelle Belegung rot. Das ist kein fertiges Produkt. Mich interessiert: Habt ihr ein No-show-Problem – und würde ein sichtbarer Check-in am Raum helfen?

**Begleitpost:**

> Ein gebuchter Raum ist nicht immer ein genutzter Raum.\
>
> In meinem **eigenen POC** mit PuKK Workspace und 3V Rooms zeigt Orange einen Termin, der noch Check-in braucht. Ein NFC-Tap bestätigt ihn, der Ring wird Rot.\
>
> Das ist keine offizielle Integration und nicht produktreif. Ich teste, ob diese Interaktion ein reales Problem trifft.\
>
> **Habt ihr ein No-show-Problem bei Meetingräumen? Würde ein sichtbarer Check-in am Raum helfen – oder nur mehr Hardware erzeugen?**\
>
> Produktkontext: [PuKK Workspace von proDVX](https://www.prodvx.com/de/produkte/7031200-pukk-workspace).

## Video 2: Buchen und Verlängern

**Hook:** „Ein Druck. Blau. Die Zeit ist noch nicht gebucht – aber eindeutig ausgewählt.“

**Hypothese:** Eine physische Auswahl direkt am Raum reduziert Reibung beim spontanen Buchen oder Verlängern, ohne unklare Berechtigungen zu erzeugen.

| Zeit | Bild und Aktion | Sprechertext / Untertitel |
| --- | --- | --- |
| 0:00–0:03 | Nahaufnahme: freie grüne LEDs werden nach Tastendruck nacheinander Blau. | Oben: „EIGENER POC – nicht produktreif“<br>„Zeit am Raum auswählen?“ |
| 0:03–0:10 | Roter aktueller Termin, dann Tastendruck; freie Anschlusssegmente werden Blau. | „Bei einer laufenden Buchung wählt die Taste freie Zeit zum Verlängern vor.“ |
| 0:10–0:16 | Schnitt auf bestätigten roten Zustand. | Klein: „Bestätigung nach 5 s“<br>„Danach übergibt der POC die Verlängerung.“ |
| 0:16–0:23 | Neuer kurzer Take: Ring vollständig Grün, ein Druck, erste Segmente Blau, dann Rot. | „Ohne aktuellen Termin wird dieselbe Auswahl im POC zur Ad-hoc-Buchung.“ |
| 0:23–0:30 | Ruhige Gesamtaufnahme des Rings. | „Wäre das bei euch hilfreich? Und wer dürfte am Raum buchen oder verlängern?“ |

**Voice-over:**

> Ein Druck. Blau. Die Zeit ist noch nicht gebucht, aber eindeutig ausgewählt. Bei einer laufenden Buchung wählt die Taste im POC freie Zeit zum Verlängern vor. Nach fünf Sekunden wird die Änderung übergeben. Ist kein Termin aktiv, funktioniert dieselbe Auswahl als Ad-hoc-Buchung. Wäre das bei euch hilfreich? Und wer dürfte am Raum buchen oder verlängern?

**Begleitpost:**

> Kalender können Räume buchen. Aber wie viel Reibung entsteht, wenn jemand im Raum noch 15 Minuten braucht?\
>
> Mein POC nutzt die PuKK-Taste für zwei Varianten:\
> - Bei laufender Buchung: freie Zeit auswählen und verlängern.\
> - Ohne laufende Buchung: freie Zeit ad hoc buchen.\
>
> Blau bedeutet zunächst nur: Auswahl sichtbar bestätigt. Erst nach fünf Sekunden ohne weitere Eingabe übergibt der POC die Änderung.\
>
> **Wäre das bei euch hilfreich? Und wichtiger: Wer dürfte am Raum buchen oder verlängern?**\
>
> Eigener POC, keine offizielle proDVX- oder 3V-Rooms-Integration. [PuKK Workspace](https://www.prodvx.com/de/produkte/7031200-pukk-workspace).

## Video 3: Terminieren

**Hook:** „Der Termin endet früher. Wann wird der Raum wieder wirklich frei?“

**Hypothese:** Frühzeitiges Terminieren hat genug praktischen Wert, um eine bewusste physische Interaktion am Raum zu rechtfertigen.

| Zeit | Bild und Aktion | Sprechertext / Untertitel |
| --- | --- | --- |
| 0:00–0:03 | Extreme Nahaufnahme: Rot wird in einer blauen Rücklauf-Animation zu Grün. | Oben: „EIGENER POC – nicht produktreif“<br>„Früher fertig?“ |
| 0:03–0:08 | Roter Ring, Hand legt sich auf die Taste. | „Der Raum ist noch reserviert, aber das Meeting ist vorbei.“ |
| 0:08–0:12 | Hand hält die Taste, dann loslassen. | Klein: „3 s halten, dann loslassen“ |
| 0:12–0:20 | Erst nach dem Loslassen: blaue Rücklauf-Animation, anschließend Grün. | „Danach terminiert der POC die aktuelle Belegung und macht den Raum wieder frei.“ |
| 0:20–0:26 | Ruhiger grüner Ring. | „Passiert das bei euch oft genug, dass diese Interaktion wertvoll wäre?“ |

**Voice-over:**

> Der Termin endet früher. Wann wird der Raum wieder wirklich frei? In meinem POC hält man die Taste drei Sekunden und lässt sie los. Erst dann erhält der Server das Langdruck-Ereignis. Der Ring läuft Blau zurück und wird nach erfolgreichem Terminieren Grün. Passiert das bei euch oft genug, dass so eine Interaktion wertvoll wäre?

**Begleitpost:**

> Meeting früher beendet, Raum bleibt aber noch eine halbe Stunde blockiert: Passiert das bei euch?\
>
> Mein POC testet dafür eine bewusste Aktion: Taste drei Sekunden halten, loslassen, aktuelle Belegung terminieren. Die LEDs laufen Blau zurück und werden Grün, wenn der Raum wieder frei ist.\
>
> Die Animation startet technisch erst nach dem Loslassen; so meldet das Gerät den Langdruck an den Server.\
>
> **Wie löst ihr frühe Meeting-Enden heute? Und wäre das häufig genug, um eine Interaktion direkt am Raum zu rechtfertigen?**\
>
> Eigener POC, nicht produktreif. [PuKK Workspace von proDVX](https://www.prodvx.com/de/produkte/7031200-pukk-workspace).

## Optionales Video 4: Weitere Orte

Dieses Video gibt es nur, wenn die ersten drei Folgen klare Resonanz für den Raumprozess zeigen. Es öffnet die Frage nach dem zweiten Einsatzort, ohne Funktionen vorzutäuschen:

> Der gezeigte Raumablauf ist mein aktueller POC. proDVX nennt auch Desk Booking, Office Pods, Statusanzeigen, gemeinsame Ressourcen und Housekeeping als mögliche Anwendungsfelder. Welche konkrete Situation wäre bei euch als Nächstes interessant?

Desk, Reinigung, Pflege oder Asset-Management dürfen nicht als aktuelle POC-Funktionen dargestellt werden. Es sind Hypothesen beziehungsweise von proDVX genannte Anwendungsfelder.

## Veröffentlichung und Auswertung

- Video 1 zuerst veröffentlichen, dann vier bis sieben Tage Kommentare auswerten.
- Video 2 nur mit derselben Tonalität nachlegen, aber ohne auf eine vorherige Folge zu verweisen.
- Video 3 als Abschluss der Kernserie veröffentlichen.
- Nicht nur „Danke“ schreiben: nach Ort, Häufigkeit, heutigem Workaround, Berechtigung und schlimmstem Fehlerfall fragen.
- Auf starke Kommentare mit einem 20–30-Sekunden-Reply reagieren, statt sofort ein weiteres allgemeines Video zu produzieren.
- proDVX oder 3V Rooms nur markieren, wenn die Kommunikation abgestimmt ist. Ein Tag kann wie eine Partnerschafts- oder Freigabeaussage wirken.

| Feld für die Auswertung | Beispiel |
| --- | --- |
| Rolle und Segment | Facility Manager, Coworking-Betreiber, AV-Integrator |
| Folge / Funktion | Check-in, Buchen & Verlängern, Terminieren |
| Ort und Situation | Meetingraum, Termin endet früh, nächster Gast wartet |
| Heutiger Workaround | Kalender, Schild, Zuruf, App |
| Einwand oder Risiko | „Bei uns darf nicht jede Person verlängern.“ |
| Nächste Handlung | Rückfrage, 20-Minuten-Call, verwerfen |

Nach der dritten Folge nicht nur Aufrufe zählen. Priorisiere wiederkehrende Probleme, konkrete Arbeitsabläufe und Gesprächsanfragen. Ein präziser negativer Einwand ist ein Erfolg, weil er die nächste Produktentscheidung schärft.

## Quellen und Referenzen

1. [proDVX: PuKK Workspace (R26-8281)](https://www.prodvx.com/de/produkte/7031200-pukk-workspace) – LED-Ring, Taste, NFC, WLAN/Bluetooth, REST-API, Maße und Einsatzfelder.
2. [proDVX: Getting started with the PuKK Workspace](https://www.prodvx.com/de/getting-started-with-the-pukk-workspace) – Konfiguration, REST-API und praktische Anwendungsfelder.
3. [ProDVX: PuKK Overview Demo-Repository](https://github.com/ProDVX/PuKK_Overview) – offizielle Demo zur Serveranbindung und LED-Ansteuerung.
4. [ProDVX: PuKK Remote Server API](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_Event_Specification.html) – dokumentierte Button- und NFC-Ereignisse.
5. [ProDVX: PuKK REST API Specification](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_API_Specification.html) – individuelle Ansteuerung der 12 LEDs.
6. [LinkedIn: How to Create Engaging Videos for Your LinkedIn Feed](https://www.linkedin.com/business/talent/blog/talent-acquisition/how-create-videos-for-linkedin-feed) – Hook, Hochformat, Untertitel und Serienproduktion.
7. [LinkedIn: B2B-Video-Tipps](https://www.linkedin.com/business/marketing/blog/content-marketing/13-top-tips-for-compelling-b2b-video-content-on-linkedin) – erste Sekunden, kurze Videos und Wiederverwertung.
