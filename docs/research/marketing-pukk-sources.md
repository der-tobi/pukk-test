# Quellenbasis: PuKK-LinkedIn-PoC

Stand: 2026-09-09. Diese Notiz trennt belegte Produktfakten von Aussagen über
den lokalen Demonstrator. Sie ist die Faktenbasis für Video-Skript, Post-Text
und Kommentare, keine Freigabe durch ProDVX oder 3V Rooms.

## Aussagen, die im Video sicher sind

### ProDVX PuKK Workspace

- Das Produkt heißt **PuKK Workspace (R26-8281)**. ProDVX bezeichnet es als
  interaktives Buchungsgerät mit Full-RGB-LED-Ring, Interaktionsknopf,
  NFC-Modul, WLAN, Bluetooth und REST API. Das ist eine gute, sachliche
  Produktvorstellung im Abspann oder Begleittext. Quelle:
  [ProDVX Produktseite](https://www.prodvx.com/de/produkte/7031200-pukk-workspace).
- Laut derselben Produktseite misst das Gerät 80 x 80 x 20,6 mm, wird über
  USB-C mit Strom versorgt und besitzt einzeln steuerbare Full-RGB-LEDs. Für
  das Bild darf daher gesagt werden: "ein kleiner, farbiger Statusindikator".
  Nicht ohne eigene Messung behaupten: konkrete Sichtweite, Akkulaufzeit,
  Installationsaufwand oder Zuverlässigkeit im Feld. Quelle:
  [ProDVX Produktseite](https://www.prodvx.com/de/produkte/7031200-pukk-workspace).
- ProDVX nennt unter anderem Desk Booking, Bestätigung einer
  Meetingraum-Buchung, Office-Pod-Booking, Statusanzeigen, gemeinsame
  Ressourcen, Nurse Calling sowie Asset-/Bed-Management als Anwendungsfelder.
  Diese Liste eignet sich als offene Frage an die Zielgruppe, nicht als
  Kundenreferenz oder als Beleg für fertige Integrationen. Quelle:
  [PuKK Workspace: Getting started](https://www.prodvx.com/de/getting-started-with-the-pukk-workspace).
- ProDVX stellt für das Gerät Konfiguration per NFC/Companion App,
  Fernsteuerung über REST API und OTA als dokumentierte Themen bereit. Das
  stützt die Aussage, dass die Hardware für eine Software-Integration
  vorgesehen ist. Quelle:
  [PuKK Workspace: Getting started](https://www.prodvx.com/de/getting-started-with-the-pukk-workspace).

### Interaktionen, die der PoC sichtbar machen kann

- Die offizielle Remote-Server-Spezifikation dokumentiert Button-Ereignisse
  (kurz, doppelt, mehrfach sowie 3-, 5- und 15-Sekunden-Long-Press) und ein
  NFC-Ereignis. Daher kann das Video sauber formulieren: "Die Hardware liefert
  Button- und NFC-Ereignisse an eine Server-Integration." Quelle:
  [PuKK Remote Server API](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_Event_Specification.html).
- Die offizielle REST-Spezifikation dokumentiert individuell steuerbare Farben
  für genau 12 LEDs sowie eine lokale REST API des Geräts. Daher ist eine
  farbliche 12-Segment-Zeit-/Statusanzeige technisch plausibel. Formulierung
  mit sauberer Grenze: "In diesem PoC nutze ich die 12 LEDs als Anzeige für
  den nächsten Zeitraum"; nicht: "ProDVX liefert diese Logik standardmäßig."
  Quelle:
  [PuKK REST API Specification](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_API_Specification.html).
- Das offizielle ProDVX-Demo-Repository beschreibt eine anpassbare
  Demo-Integration. Es darf als Inspirations- und Integrationsbeleg verlinkt
  werden, aber nicht als produktionsreife oder offizielle 3V-Rooms-Referenz.
  Quelle: [ProDVX PuKK Overview](https://github.com/ProDVX/PuKK_Overview).

### Der lokale 3V-Rooms-PoC

- Sicher als eigene, nicht als Hersteller-Aussage: "Das ist ein persönlicher
  PoC, keine Produktankündigung und keine freigegebene Standardintegration."
- Sicher, sofern genau der gezeigte Stand gefilmt wird: "Mein PoC liest
  Belegungsdaten aus 3V Rooms und stellt einen rollierenden 60-Minuten-Status
  auf dem PuKK dar." Die Implementierung und ihre Tests liegen im Repository;
  dies ist kein von 3V Rooms oder ProDVX bestätigtes Leistungsversprechen.
- Sichere Demonstrationsaussagen, nur wenn sie im Dreh gefilmt werden:
  * freie Zeit wird grün dargestellt;
  * eine aktuelle, noch nicht eingecheckte Buchung orange;
  * eine eingecheckte aktuelle Buchung rot;
  * eine spätere Buchung violett;
  * eine vorläufige Auswahl blau.
  Diese Farben und Abläufe sind Entscheidungen dieses PoCs, nicht Teil der
  ProDVX-Produktspezifikation.
- Nicht behaupten, bevor es am realen Tenant/Gerät durchgetestet ist:
  dass NFC bei jedem Mandanten eine Buchung eincheckt, dass eine Buchung stets
  verlängert werden kann, oder dass die Anzeige live/genau/echtzeitfähig ist.
  Das sind Integrations- und Berechtigungsfragen, keine Hardware-Fakten.

## Quellengebundene LinkedIn-Leitplanken

- Die ersten 4--6 Sekunden sollen mit einem Insight, einer Frage oder einem
  auffälligen Bild beginnen; Branding kommt später. Quelle:
  [LinkedIn: Lights, Camera, LinkedIn](https://www.linkedin.com/business/marketing/blog/marketing-collective/how-to-create-high-impact-video-with-a-linkedin-expert).
- LinkedIn empfiehlt für mobile Betrachtung Hochformat und Untertitel; für
  Awareness-/Consideration-Videos nennt LinkedIn weniger als 30 Sekunden und
  einen Hook innerhalb der ersten zehn Sekunden. Quelle:
  [LinkedIn: Content-Ideen für Pages](https://www.linkedin.com/business/marketing/blog/linkedin-pages/10-content-ideas-for-linkedin-page).
- LinkedIn nennt als starke B2B-Videoelemente auffällige erste Bilder,
  Untertitel und schnelle Schnitte; als Orientierung für Videoanzeigen werden
  etwa 15 Sekunden genannt. Quelle:
  [LinkedIn: Guide to ad types](https://www.linkedin.com/business/marketing/blog/linkedin-ads/a-b2b-marketers-guide-to-every-linkedin-ad-type).
- Für einen organischen Video-Upload akzeptiert LinkedIn 3 Sekunden bis
  15 Minuten, 10--60 fps und ein Seitenverhältnis zwischen 1:2,4 und 2,4:1;
  Thumbnail und Untertitel sind wählbar. Das macht 9:16 oder 4:5 technisch
  möglich. Bei der Aufnahme trotzdem wichtige Schrift und das Gerät nicht
  bis an die Bildkante setzen, weil die LinkedIn-Oberfläche Elemente
  überlagern kann. Quelle:
  [LinkedIn Help: Videos posten](https://www.linkedin.com/help/linkedin/answer/a7174587).
- Falls ein Beitrag später gezielt beworben wird, nennt LinkedIn 4:5 als
  empfohlenes vertikales Anzeigenformat sowie 15--30 Sekunden als Empfehlung
  für eine breite Platzierung. Das ist Anzeigen-, nicht organische
  Posting-Guidance. Quelle:
  [LinkedIn Help: Video-Ad-Spezifikationen](https://www.linkedin.com/help/linkedin/answer/a424737).
- Für einen rein organischen Post sind diese Quellen keine Algorithmus- oder
  Reichweitengarantie, teilweise beziehen sie sich auf Anzeigen. Sie geben
  dennoch eine belastbare kreative Richtung: ein kurzer visueller Teaser plus
  eine zweite, etwas erklärende Demonstration ist sinnvoller als ein
  ungeschnittenes 15-Minuten-Timelapse.

## Empfohlene Formulierungen

**Hook, zulässig:** "Würdet ihr einen freien Meetingraum lieber sehen als in
einer App suchen?" / "Was würdet ihr mit 12 LEDs, einem Knopf und NFC in eurem
Büro machen?"

**PoC-Hinweis, sichtbar und gesprochen:** "PoC: Eigene Demo mit PuKK Workspace
und 3V Rooms. Keine Standardintegration, kein Produktrelease."

**CTA, zulässig:** "Welchen Use Case seht ihr: Meetingraum, Flex Desk,
Telefonbox, Pflege, Reinigung oder etwas ganz anderes?"

**Nicht verwenden:** "die PuKK-3V-Rooms-Integration", "offizielle
Partnerschaft", "produktionsreif", "Echtzeit", "spart nachweislich Kosten"
oder Aussagen über Kundenwirkung ohne belegte Kundendaten bzw. Freigabe.

## Direkte Quellenliste

1. [ProDVX PuKK Workspace Produktseite](https://www.prodvx.com/de/produkte/7031200-pukk-workspace)
2. [ProDVX PuKK Workspace: Getting started](https://www.prodvx.com/de/getting-started-with-the-pukk-workspace)
3. [ProDVX PuKK Remote Server API](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_Event_Specification.html)
4. [ProDVX PuKK REST API Specification](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_API_Specification.html)
5. [LinkedIn: Lights, Camera, LinkedIn](https://www.linkedin.com/business/marketing/blog/marketing-collective/how-to-create-high-impact-video-with-a-linkedin-expert)
6. [LinkedIn: 15 Inspiring Content Ideas for Your LinkedIn Page](https://www.linkedin.com/business/marketing/blog/linkedin-pages/10-content-ideas-for-linkedin-page)
7. [LinkedIn: A B2B Marketer's Guide to Every LinkedIn Ad Type](https://www.linkedin.com/business/marketing/blog/linkedin-ads/a-b2b-marketers-guide-to-every-linkedin-ad-type)
8. [LinkedIn Help: Videos posten](https://www.linkedin.com/help/linkedin/answer/a7174587)
9. [LinkedIn Help: Video-Ad-Spezifikationen](https://www.linkedin.com/help/linkedin/answer/a424737)
