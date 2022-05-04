// +build !ci

#import <Foundation/Foundation.h>
#if __MAC_OS_X_VERSION_MAX_ALLOWED >= 101400
#import <UserNotifications/UserNotifications.h>
#endif

static int notifyNum = 0;

//extern void fallbackSend(char *cTitle, char *cBody);

bool isBundled2() {
    return [[NSBundle mainBundle] bundleIdentifier] != nil;
}

@interface MyClass:NSObject<UNUserNotificationCenterDelegate>
// ...
@end

@implementation MyClass
- (void)userNotificationCenter:(UNUserNotificationCenter *)center 
       willPresentNotification:(UNNotification *)notification 
         withCompletionHandler:(void (^)(UNNotificationPresentationOptions options))completionHandler{
             NSLog(@"received notification request,");
             completionHandler(UNNotificationPresentationOptionBanner|UNNotificationPresentationOptionSound);
         }

         - (void)userNotificationCenter:(UNUserNotificationCenter *)center 
didReceiveNotificationResponse:(UNNotificationResponse *)response 
         withCompletionHandler:(void (^)(void))completionHandler {
             NSLog(@"got action %@", response.notification.request.identifier);
             completionHandler();
         }
@end


#if __MAC_OS_X_VERSION_MAX_ALLOWED >= 101400
void doSendNotification2(UNUserNotificationCenter *center, NSString *title, NSString *body) {
    UNMutableNotificationContent *content = [UNMutableNotificationContent new];
    [content autorelease];
    content.title = title;
    content.body = body;

    notifyNum++;
    NSString *identifier = [NSString stringWithFormat:@"fyne-notify-%d", notifyNum];
    UNNotificationRequest *request = [UNNotificationRequest requestWithIdentifier:identifier
        content:content trigger:[UNTimeIntervalNotificationTrigger triggerWithTimeInterval: 1 repeats:false]];
    

    [center addNotificationRequest:request withCompletionHandler:^(NSError * _Nullable error) {
        if (error != nil) {
            NSLog(@"Could not send notification: %@", error);
        }
    }];
}

void sendNotification2(char *cTitle, char *cBody) {
    UNUserNotificationCenter *center = [UNUserNotificationCenter currentNotificationCenter];
    NSString *title = [NSString stringWithUTF8String:cTitle];
    NSString *body = [NSString stringWithUTF8String:cBody];

    MyClass *instanceOfMyClass = [[MyClass alloc] init];
    center.delegate = instanceOfMyClass;

    UNAuthorizationOptions options = UNAuthorizationOptionAlert;
    [center requestAuthorizationWithOptions:options
        completionHandler:^(BOOL granted, NSError *_Nullable error) {
            if (!granted) {
                if (error != NULL) {
                    NSLog(@"Error asking for permission to send notifications %@", error);
                    // this happens if our app was not signed, so do it the old way
                    //fallbackSend((char *)[title UTF8String], (char *)[body UTF8String]);
                } else {
                    NSLog(@"Unable to get permission to send notifications");
                }
            } else {
                NSLog(@"auth granted");
                doSendNotification2(center, title, body);
            }
        }];
}
#else
void sendNotification2(char *cTitle, char *cBody) {
	//fallbackSend(cTitle, cBody);
}
#endif
