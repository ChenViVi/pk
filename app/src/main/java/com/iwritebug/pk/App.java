package com.iwritebug.pk;

import android.app.Application;

import com.mcxiaoke.packer.helper.PackerNg;
import com.umeng.commonsdk.UMConfigure;

public class App extends Application {

    @Override
    public void onCreate() {
        super.onCreate();
        UMConfigure.init(this, "5b795768f29d980ee8000871", PackerNg.getChannel(this), UMConfigure.DEVICE_TYPE_PHONE, null);
    }
}
