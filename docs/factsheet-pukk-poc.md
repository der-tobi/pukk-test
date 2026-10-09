# Factsheet: PuKK Workspace mit 3V Rooms

Stand: 9. Oktober 2026 · Grundlage für einen Blogpost über den aktuellen PoC.

## Idee und Grundlagen

Der Demonstrator verbindet den **proDVX PuKK Workspace** mit **3V Rooms**:
Raumbelegung wird als farbiger LED-Ring sichtbar; Buchung, Verlängerung,
Check-in und Freigabe lassen sich direkt am Gerät auslösen.

- **Hardware:** 12 einzeln steuerbare RGB-LEDs, ein Taster, NFC-Leser und WLAN.
- **Integration:** Ein eigener Go-Server vermittelt zwischen dem PuKK und der
  Rooms-API. Der PuKK liefert Tastendruck-/NFC-Ereignisse; der Server steuert
  seine LEDs über die lokale REST-API.
- **Betrieb:** Ein einzelnes ausführbares Programm, auch als Windows-EXE ohne
  zusätzliche Laufzeitumgebung. Keine Datenbank, kein separates Frontend.
- **PoC-Konfiguration:** Ein Gerät und eine fest hinterlegte Rooms-Ressource.
  Das API-Passwort wird beim Start abgefragt. Gerät und Server müssen sich
  im Netzwerk erreichen können; der Server benötigt Zugriff auf Rooms.

## Aktueller Funktionsumfang

| Funktion | Verhalten im PoC |
| --- | --- |
| Zeitübersicht | Rollierende nächste **60 Minuten**, dargestellt durch **12 × 5 Minuten**. Laufende und kommende Buchungen sowie freie Zwischenräume werden sichtbar. |
| Spontan buchen | Im freien Raum fügt ein kurzer Tastendruck bis zu **15 Minuten** hinzu. Weitere Tastendrücke erweitern die Auswahl. |
| Verlängern | Bei laufender Buchung erweitert derselbe Taster deren Ende. Grenze: nächste bekannte Buchung oder eine Stunde ab jetzt. Auch ein Restfenster von beispielsweise 23 Minuten lässt sich exakt ausnutzen. |
| Auswahl zurücknehmen | Nach Erreichen der Grenze entfernen weitere Tastendrücke die hinzugefügten Schritte rückwärts. |
| Speichern | Nach **5 Sekunden ohne weiteren Tastendruck** wird die Auswahl in Rooms gespeichert. Eine am Gerät neu erstellte Buchung wird bei offenem Check-in direkt eingecheckt. |
| NFC-Check-in | Ein NFC-Tap bestätigt eine zum Check-in bereite Buchung. Ein von Rooms geliefertes Vorab-Check-in-Fenster wird berücksichtigt. |
| Vorzeitig freigeben | Die Longpress-Ereignisse **3, 5 und 15 Sekunden** beenden die laufende Buchung und geben den Raum ohne Nachlaufzeit frei. |
| Rückmeldung | Auswahl und Verlängerung erhalten blaue LED-Rückmeldungen. Beim Terminieren zeigt ein statischer blauer Zustand die Verarbeitung an; anschließend erscheint die normale Belegung. |
| Marketing-Startsequenz | Beim ersten Gerätekontakt je Serverstart leuchtet der Ring fünf Sekunden konstant in gesättigten Referenzfarbtönen mit abgestimmter Helligkeit. Danach erscheint die normale Belegung. |
| Diagnose | Server-IP und Anfragen im Terminal; Erreichbarkeitsprüfung und Statusdiagnose über `/healthz` und `/debug/status`. |

### Farben dieses Demonstrators

| Farbe | Bedeutung |
| --- | --- |
| Grün | Freie Zeit |
| Rot | Belegte Zeit, insbesondere die laufende/eingecheckte Buchung |
| Orange | Buchung ist zum Check-in bereit, aber noch nicht eingecheckt |
| Violett | Weitere beziehungsweise Folgebuchungen |
| Blau | Vorläufige Auswahl oder Verarbeitung einer Freigabe |

## Aktualisierung und Grenzen

Rooms-Daten werden im Hintergrund alle **30 Sekunden** aktualisiert und
zwischengespeichert. Der PuKK pollt normalerweise alle **5 Sekunden**.
Bereits geladene Terminstarts und Check-in-Zeitfenster werden bei jedem Poll
neu bewertet. Die Anzeige hat fünf Minuten Auflösung; Buchungsgrenzen
werden anhand der exakten Terminzeiten berechnet.

Der PoC akzeptiert jede NFC-Karte; sie dient als Auslöser, nicht zur
Identitätsprüfung. Aktionen erfolgen mit den Berechtigungen des Rooms-API-Nutzers.
Bei fehlenden Belegungsdaten werden unbekannte Zeitabschnitte als belegt
angezeigt. Scheitert nur der automatische Check-in, bleibt die erstellte
Buchung erhalten und kann per NFC eingecheckt werden.

Die Firmware erzeugt beim Tastendruck eine eigene hellgelbe/weiß wirkende
Rückmeldung. Deren Abschaltung ist nicht dokumentiert. Ein verlässliches
„Jetzt loslassen“-Signal genau nach drei Sekunden während des Haltens ist
mit den dokumentierten Geräteereignissen nicht zugesichert.

## Einordnung für den Blogpost

**Eigener Proof of Concept**, keine freigegebene Standardintegration oder
Produktankündigung. Farben und Buchungslogik stammen aus diesem Projekt.
Der aktuelle Code wurde mit automatisierten Tests einschließlich
Nebenläufigkeitsprüfung und Windows-Build geprüft; die jüngsten Änderungen
benötigen noch Prüfung am echten Gerät. Echtzeit, Produktionsreife oder
nachgewiesene Zeit-/Kostenersparnis sind keine belegten Aussagen.

Möglicher Blickwinkel: „Wie sich Raumbelegung und einfache Buchungsaktionen
mit zwölf LEDs, einem Taster und NFC direkt am Raum zugänglich machen lassen.“

## Quellen zur Vertiefung

- [ProDVX: Hardware, Ereignisse und REST-API](https://prodvx.github.io/docs/PuKK_Workspace/)
- [3V Rooms: API-Dokumentation](https://docs.3vrooms.app/3vrooms/api/)
- [Projektstand und Implementierungsdetails](../memory/poc_implementation.md)
- [Gesammelte Produktquellen und Hinweise für öffentliche Aussagen](research/marketing-pukk-sources.md)
