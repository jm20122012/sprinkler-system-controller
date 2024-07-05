#include <WiFi.h>
#include <NTPClient.h>
#include <WiFiUdp.h>
#include <time.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>

struct Zone {
  const char* zoneName;
  int pin;
  unsigned long lastOnTime;
  bool isActive;
};

// Constants
// const int LED = 25;
const int ZONE_COUNT = 4;
const int ZONE_1_PIN = 6;
const int ZONE_2_PIN = 7;
const int ZONE_3_PIN = 8;
const int ZONE_4_PIN = 9;

Zone zoneList[ZONE_COUNT] = {
  {"zone1", ZONE_1_PIN, 0, false},
  {"zone2", ZONE_2_PIN, 0, false},
  {"zone3", ZONE_3_PIN, 0, false},
  {"zone4", ZONE_4_PIN, 0, false}
};

const char* ssid = "";
const char* password = "";
const char* mqtt_broker = "";
const int mqtt_port = 1883;

const unsigned long CONNECTION_CHECK_INTERVAL = 10; // In seconds
const unsigned long NTP_UPDATE_INTERVAL = 60; // In seconds
const unsigned long MAX_ZONE_ON_TIME = 60 * 20; // Max on time in seconds for any zone
const unsigned long LED_BLINK_INTERVAL = 1000; // In milliseconds
const unsigned long PUBLISH_STATUS_INTERVAL = 3; // in seconds

const char* MQTT_CMD_TOPIC = "sprinkler_system_controller/picow/command";
const char* MQTT_STATUS_TOPIC = "sprinkler_system_controller/picow/status";
const char* MQTT_RESPONSE_TOPIC = "sprinkler_system_controller/picow/response";

// Global variables
unsigned long lastConnectionCheck = 0;
unsigned long lastLedBlink = millis();
unsigned long lastStatusPub = 0;

bool ledState = false;

WiFiUDP ntpUDP;
NTPClient timeClient(ntpUDP, "pool.ntp.org", 0, NTP_UPDATE_INTERVAL);

WiFiClient mqttClient;
PubSubClient client(mqttClient);

// Function declarations
void connectWifi();
String getUTCTimestamp();
void log(String msg);
void mqttReconnect();
void mqttMSgCallback(char* topic, byte* paylod, unsigned int length);
void updateZoneState(String zoneStr, int state);
void updateZoneState(int zoneIndex, bool newState);
void cleanup();
void publishStatus();

void setup() {
  Serial.begin(115200);
  while (!Serial) {
    ; // wait for serial port to connect. Needed for native USB port only
  }

  log("Starting setup...");
  
  // Set all zone pins to initial state of off
  cleanup();

  // Wifi Setup
  connectWifi();
  
  // NTP Setup
  timeClient.begin();
  timeClient.setTimeOffset(0);
  if (!timeClient.update()){
    log("NTP update failed");
  }

  // MQTT Setup
  client.setServer(mqtt_broker, mqtt_port);
  client.setCallback(mqttMsgCallback);
  client.setBufferSize(1024);

  // GPIO Setup
  pinMode(LED_BUILTIN, OUTPUT);
  pinMode(ZONE_1_PIN, OUTPUT);
  pinMode(ZONE_2_PIN, OUTPUT);
  pinMode(ZONE_3_PIN, OUTPUT);
  pinMode(ZONE_4_PIN, OUTPUT);

  log("Setup complete.");
}

void loop() {
  if (!client.connected()) {
    cleanup();
    mqttReconnect();
  }

  client.loop();
  unsigned long currentTime = timeClient.getEpochTime();

  if (currentTime - lastConnectionCheck >= CONNECTION_CHECK_INTERVAL){
    log("Checking WiFi connection...");
    connectWifi();
    lastConnectionCheck = currentTime;
  }

  for (int i = 0; i < ZONE_COUNT; i++){
    Zone& zone = zoneList[i];
    if (zone.isActive && currentTime - zone.lastOnTime >= MAX_ZONE_ON_TIME){
      log("Zone " + String(zone.zoneName) + " has exceeded MAX_ZONE_ON_TIME of " + String(MAX_ZONE_ON_TIME) + "s");
      updateZoneState(i, false);
    }
  }

  if (currentTime - lastStatusPub >= PUBLISH_STATUS_INTERVAL){
    publishStatus();
    lastStatusPub = currentTime;
  }

  if (millis() - lastLedBlink >= LED_BLINK_INTERVAL){
    ledState = !ledState;
    digitalWrite(LED_BUILTIN, ledState);
    lastLedBlink = millis();
  }
}

void connectWifi() {
  if (WiFi.status() != WL_CONNECTED) {
    log("WiFi not connected. Attempting to connect...");
    cleanup();
    WiFi.disconnect();
    WiFi.begin(ssid, password);
    
    int attempts = 0;
    while (WiFi.status() != WL_CONNECTED && attempts < 20) {
      delay(500);
      attempts++;
    }
    if (WiFi.status() == WL_CONNECTED) {
      log("Connected to WiFi");
      IPAddress ip = WiFi.localIP();
      char ipStr[16];
      snprintf(ipStr, sizeof(ipStr), "%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3]);
      log("IP address: " + String(ipStr));
    } else {
      log("Failed to connect to WiFi");
    }
  } else {
    log("WiFi connection is stable");
  }
}

String getUTCTimestamp() {
  time_t epochTime = timeClient.getEpochTime();
  struct tm *ptm = gmtime(&epochTime);
  
  char buffer[25];
  strftime(buffer, sizeof(buffer), "%Y-%m-%d %H:%M:%S UTC", ptm);
  
  return String(buffer);
}

void log(String msg){
  String timestamp = getUTCTimestamp();
  Serial.print("[");
  Serial.print(timestamp);
  Serial.print("] ");
  Serial.println(msg);
}

void mqttReconnect() {
  int attempts = 0;

  while (!client.connected()) {
    log("Attempting MQTT connection...");
    String clientId = "sprinklerMicrocontroller-" + String(random(0xffff), HEX);
    if (client.connect(clientId.c_str())) {
      log("MQTT Connected");

      if (client.subscribe(MQTT_CMD_TOPIC, 1)){
        log("Subscribed to topic: " + String(MQTT_CMD_TOPIC));
      } else {
        log("Failed to subscribe to topic: " + String(MQTT_CMD_TOPIC));
      }
    } else {
      attempts++;
      Serial.print("failed, rc=");
      Serial.print(client.state());
      Serial.println(" try again in 5 seconds");
      delay(5000);
    }
    if (attempts >= 5){
      return;
    }
  }
}

void mqttMsgCallback(char* topic, byte* payload, unsigned int length) {
  char message[length + 1];
  memcpy(message, payload, length);
  message[length] = '\0';
  
  char logMessage[300];
  snprintf(logMessage, sizeof(logMessage), "MQTT Msg Received - Topic: %s - Msg: %s", topic, message);
  log(String(logMessage));

  JsonDocument doc;  // Allocate a JsonDocument

  // Deserialize the JSON document
  DeserializationError error = deserializeJson(doc, payload, length);
  if (error) {
    log("JSON parsing failed: " + String(error.c_str()));
    return;
  }

  if (strcmp(topic, MQTT_CMD_TOPIC) == 0) {
    String commandType = doc["command_type"];

    if (commandType == "update_zone_state") {
      updateZoneState(doc["zone"].as<String>(), doc["state"].as<int>());
    } else {
      log("Unknown command type received: " + commandType);
    }
  }
}

void updateZoneState(String zoneStr, int state) {
    int zoneIndex = -1;
    if (zoneStr == "zone1") zoneIndex = 0;
    else if (zoneStr == "zone2") zoneIndex = 1;
    else if (zoneStr == "zone3") zoneIndex = 2;
    else if (zoneStr == "zone4") zoneIndex = 3;
    else {
        log("Unknown zone: " + zoneStr);
        return;
    }

    if (state != 0 && state != 1) {
      log("Unknown state received in command: " + String(state));
      return;
    }

    if (state == 1){
      for (int i = 0; i < ZONE_COUNT; i++){
        Zone& zone = zoneList[i];

        if (digitalRead(zone.pin) == HIGH){
          log("Zone ON cmd recieved for " + zoneStr + " but zone " + String(i + 1) + " is currently ON");
          return;
        }
      }
    }

    updateZoneState(zoneIndex, state == 1);
}

void updateZoneState(int zoneIndex, bool newState){
  if (zoneIndex < 0 || zoneIndex >= ZONE_COUNT) {
    log("Invalid zone index: " + String(zoneIndex));
    return;
  }

  Zone& zone = zoneList[zoneIndex];

  if (zone.isActive != newState){
    zone.isActive = newState;
    digitalWrite(zone.pin, newState ? HIGH : LOW);

    if (newState) {
      unsigned long currentTime = timeClient.getEpochTime(); // In seconds
      zone.lastOnTime = currentTime;
    }
  }
}

void cleanup(){
  for (int i = 0; i < ZONE_COUNT; i++){
    updateZoneState(i, false);
  }
}

void publishStatus() {
  JsonDocument doc;

  doc["messageType"] = "status";

  unsigned long currentTime = timeClient.getEpochTime();

  for (int i = 0; i < ZONE_COUNT; i++) {
    doc[String(zoneList[i].zoneName) + "Active"] = zoneList[i].isActive;

    String lastOnTimeName = String(zoneList[i].zoneName) + "LastOnTime";
    // log("Parsing lastOnTimeName: " + lastOnTimeName);

    if (zoneList[i].lastOnTime > 0) {
      time_t lastOnTime = zoneList[i].lastOnTime;
      char timeStr[25];
      strftime(timeStr, sizeof(timeStr), "%Y-%m-%d %H:%M:%S UTC", gmtime(&lastOnTime));
      // log("Parsing last on time: " + String(timeStr));
      doc[lastOnTimeName] = timeStr;
    } else {
      doc[lastOnTimeName] = "never";
    }
  }

  char jsonBuffer[1024];  // Increased buffer size
  size_t jsonSize = serializeJson(doc, jsonBuffer);
  // log("JSON payload size: " + String(jsonSize) + " bytes");

  if (!client.connected()) {
    log("MQTT client disconnected. Attempting to reconnect...");
    cleanup();
    mqttReconnect();
  }

  // log("Attempting to publish. Client state: " + String(client.state()) + ", Connected: " + String(client.connected()));
  bool pubResult = client.publish(MQTT_STATUS_TOPIC, jsonBuffer);
  // log("Publish result: " + String(pubResult) + ", Client state after publish: " + String(client.state()));

  if (pubResult) {
    log("Status published successfully");
  } else {
    log("Failed to publish status. MQTT state: " + String(client.state()));
  }
}