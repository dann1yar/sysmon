import Clutter from 'gi://Clutter';
import GObject from 'gi://GObject';
import Gio from 'gi://Gio';
import GLib from 'gi://GLib';
import St from 'gi://St';
import * as Main from 'resource:///org/gnome/shell/ui/main.js';
import * as PanelMenu from 'resource:///org/gnome/shell/ui/panelMenu.js';
import * as PopupMenu from 'resource:///org/gnome/shell/ui/popupMenu.js';
import { Extension } from 'resource:///org/gnome/shell/extensions/extension.js';

Gio._promisify(Gio.Subprocess.prototype, 'communicate_utf8_async');

const SYSMON = GLib.get_home_dir() + '/projects/sysmon/bin/sysmon';
const INTERVAL_SEC = 3;

const ITEMS = [
    ['cpu', 'cpu.svg'],
    ['ram', 'ram.svg'],
    ['disk', 'disk.svg'],
    ['temp', 'temp.svg'],
    ['net', 'net.svg'],
];

function humanRate(bps) {
    const units = ['B', 'K', 'M', 'G'];
    let i = 0;
    while (bps >= 1024 && i < units.length - 1) { bps /= 1024; i++; }
    return `${Math.round(bps)}${units[i]}`;
}

const gib = b => (b / 1073741824).toFixed(1);

const Indicator = GObject.registerClass(
    class Indicator extends PanelMenu.Button {
        _init(extDir) {
            super._init(0.0, 'Sysmon', false);   // true = без меню

            // --- полоса на панели ---
            const box = new St.BoxLayout();
            this._labels = {};
            for (const [key, file] of ITEMS) {
                box.add_child(new St.Icon({
                    gicon: Gio.icon_new_for_string(`${extDir}/icons/${file}`),
                    icon_size: 16,
                    y_align: Clutter.ActorAlign.CENTER,
                }));
                const label = new St.Label({
                    text: '…',
                    y_align: Clutter.ActorAlign.CENTER,
                    style: 'margin-left: 4px; margin-right: 12px;',
                });
                this._labels[key] = label;
                box.add_child(label);
            }
            this.add_child(box);

            // --- меню ---
            this._rows = {};
            for (const key of ['cpu', 'ram', 'disk', 'io', 'temp', 'net']) {
                const item = new PopupMenu.PopupMenuItem('…', { reactive: false });
                this.menu.addMenuItem(item);
                this._rows[key] = item;
            }

            this._cancellable = new Gio.Cancellable();
            this._busy = false;
            this._timer = GLib.timeout_add_seconds(GLib.PRIORITY_DEFAULT, INTERVAL_SEC, () => {
                this._poll();
                return GLib.SOURCE_CONTINUE;
            });
            this._poll();
        }

        async _poll() {
            if (this._busy) return;
            this._busy = true;
            try {
                const proc = Gio.Subprocess.new(
                    [SYSMON, '--once', '--json'], Gio.SubprocessFlags.STDOUT_PIPE);
                const [out] = await proc.communicate_utf8_async(null, this._cancellable);
                this._update(JSON.parse(out));
            } catch (e) {
                if (!e.matches?.(Gio.IOErrorEnum, Gio.IOErrorEnum.CANCELLED))
                    console.error(`sysmon: ${e}`);
            } finally {
                this._busy = false;
            }
        }

        _update(d) {
            const temp = d.temp_ok ? `${Math.round(d.temp_c)}°C` : '–';
            const net = `↓${humanRate(d.net_rx_bytes_per_sec)} ↑${humanRate(d.net_tx_bytes_per_sec)}`;

            // полоса
            this._labels.cpu.text = `${Math.round(d.cpu_percent)}%`;
            this._labels.ram.text = `${Math.round(d.ram_percent)}%`;
            this._labels.disk.text = `${Math.round(d.disk_percent)}%`;
            this._labels.temp.text = temp;
            this._labels.net.text = net;

            const c = d.disk_percent >= 85 ? '#ff5555'
                : d.disk_percent >= 60 ? '#f5d142' : '';
            this._labels.disk.style = `margin-left: 4px; margin-right: 12px; color: ${c};`;

            // меню
            this._rows.cpu.label.text = `CPU:  ${d.cpu_percent.toFixed(1)}%`;
            this._rows.ram.label.text = `RAM:  ${d.ram_percent.toFixed(1)}%  (${gib(d.ram_used_bytes)} / ${gib(d.ram_total_bytes)} GiB)`;
            this._rows.disk.label.text = `Disk: ${d.disk_percent.toFixed(1)}%  (${gib(d.disk_used_bytes)} / ${gib(d.disk_total_bytes)} GiB)`;
            this._rows.io.label.text = `I/O:  R ${humanRate(d.disk_read_bytes_per_sec)}/s  W ${humanRate(d.disk_write_bytes_per_sec)}/s`;
            this._rows.temp.label.text = `Temp: ${temp}`;
            this._rows.net.label.text = `Net ${d.net_iface}:  ${net}`;
        }

        destroy() {
            this._cancellable.cancel();
            if (this._timer) GLib.Source.remove(this._timer);
            this._timer = null;
            super.destroy();
        }
    });

export default class SysmonExtension extends Extension {
    enable() {
        this._indicator = new Indicator(this.path);
        Main.panel.addToStatusArea(this.uuid, this._indicator, 0, 'right');
    }
    disable() {
        this._indicator?.destroy();
        this._indicator = null;
    }
}