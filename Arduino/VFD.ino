#include <SoftwareSerial.h>
#include <SPI.h>
#include <Wire.h>
#include <EtherCard.h>
#include <ArduinoJson.h>


// Ethernet
byte mac[] = {0xDE, 0xAD, 0xBE, 0xEF, 0xBA, 0xBE};
byte myip[] = { 192, 168, 0, 6 };
byte gwip[] = { 192, 168, 0, 1 };
byte Ethernet::buffer[512];
boolean OUTSTANDING_WWW_REQ = false;

// VFD
SoftwareSerial VFD(6,10); // RX, TX
#define VFD_LINES 4
#define VFD_COLS 20
#define VFD_COLMAX 64
char VFD_TEXT[VFD_LINES][VFD_COLMAX];
int VFD_SCROLL_INDEX[VFD_LINES];
int VFD_CUR_LINELEN[VFD_LINES];
short luminance = 0xFF;
short charset_control = 0x18;
#define VFD_TICK_MILLIS 125
#define DATA_CHECK_MILLIS 15000

// Function prototypes
void postTempData(short node_id, float temp, float pressure, float humidity);
void parseHTTPResponse();
void http_handle_resp(byte status, word off, word len);
void getInfoToShow();
void tick_vfd();

void setup() {
    // Start debug serial
    Serial.begin(9600);

    // Start serial for VFD
    VFD.begin(9600);
    int i, j;
    for (i = 0; i < VFD_LINES; i++){
        for (j=0; j < VFD_COLS; j++){
            if (i % 2){
                VFD_TEXT[i][j] = 0xD7;
            } else {
                VFD_TEXT[i][j] = '+';
            }
        }
        VFD_CUR_LINELEN[i] = VFD_COLS;
    }

    // Ethernet
    if (ether.begin(sizeof Ethernet::buffer, mac, 9) == 0){
        while (1);
    }

    ether.staticSetup(myip, gwip);
    ether.copyIp(ether.hisip, gwip);
}

/*
 * Main loop:
 * - Check for pending node transmissions
 * - Iterate over the node data we have stored, and see if any of it is dirty
 *     - If so, post it to the server
 * - Adjust relay based on response
 */
long long last_vfd_tick = 0;
long long last_data_fetch = 0;
void loop() {
    // Handle low-level ethernet data
    word len = ether.packetReceive();
    ether.packetLoop(len);

    if (millis() - last_vfd_tick > VFD_TICK_MILLIS){
        tick_vfd();
        last_vfd_tick = millis();
    }

    if (millis() - last_data_fetch > DATA_CHECK_MILLIS){
        getInfoToShow();
        last_data_fetch = millis();
    }
}

void getInfoToShow(){
    Serial.println("GET data...");
    ether.browseUrl(PSTR("/info"), "", PSTR("clock.rhye.org"), http_handle_resp);
    OUTSTANDING_WWW_REQ = 1;
    short retries = 0;
    while (OUTSTANDING_WWW_REQ && retries < 15){
        // Handle low-level ethernet data
        ether.packetLoop(ether.packetReceive());
        delay(100);
        retries++;
    }
}

void tick_vfd(){
    // Ensure flickerless mode
    VFD.write(0x1B);
    VFD.write(0x53);

    // Update luminance
    VFD.write(0x1B);
    VFD.write(0x4C);
    VFD.write(luminance);

    // Update charset
    VFD.write(charset_control);

    // Reset cursor to top left
    VFD.write(0x0C);

    // For each line, if the length is > the length of the VFD, increment the
    // scrolling tick index.
    // Print 20 characters of the line.
    int i, j, k;
    for (i=0; i < VFD_LINES; i++){
        // ESC (0x1B), 'H' (0x48), cursor position
        VFD.write(0x1B);
        VFD.write(0x48);
        VFD.write(i * 20);

        // CAN (clears the line)
        // CAN flickers a lot, use rewrite instead.
        //VFD.write(0x0F);

        if (VFD_CUR_LINELEN[i] > 20) {
            VFD_SCROLL_INDEX[i]++;
        }
        j = VFD_SCROLL_INDEX[i];
        for (k = 0; k < VFD_COLS && k < VFD_CUR_LINELEN[i]; k++){
            VFD.write(VFD_TEXT[i][j % VFD_CUR_LINELEN[i]]);
            j++;
        }
        for (; k < VFD_COLS; k++){
            VFD.write(" ");
        }
    }
}

void http_handle_resp(byte status, word off, word len){
    OUTSTANDING_WWW_REQ = 0;
    Serial.print("HTTP offset: ");
    Serial.println(off);

    Ethernet::buffer[off+len < 512 ? off+len : 511] = 0;
    for (; Ethernet::buffer[off]; off++){
        char c = Ethernet::buffer[off];
        if (c == '\n' || c == '\r'){
            off++;
            if (Ethernet::buffer[off] == '\n' || Ethernet::buffer[off] == '\r'){
                off++;
                if (Ethernet::buffer[off] == '\n' || Ethernet::buffer[off] == '\r'){
                    Serial.print("End of headers: ");
                    Serial.println(off);
                    off++;
                    break;
                }
            }
        }
    }
    Serial.println((char *)&Ethernet::buffer[off]);
    StaticJsonBuffer<512> jsonBuffer;

    JsonObject& root = jsonBuffer.parseObject((char *)&(Ethernet::buffer[off]));
    luminance = root["Luminance"];
    charset_control = root["Charset"];
    int i;
    for (i=0; i < VFD_LINES; i++){
        strncpy(VFD_TEXT[i], root["Line"][i], VFD_COLMAX);
        VFD_TEXT[i][63] = '\0';
        VFD_CUR_LINELEN[i] = strlen(VFD_TEXT[i]);
        VFD_SCROLL_INDEX[i] = 0;
        Serial.println(VFD_TEXT[i]);
    }
}
