import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import api from "../../lib/api";
import { AppDispatch } from "../store"; // Adjust the import according to your store setup

// Define the types for the state
interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

interface ConfigState {
  websiteData: any;
  aboutContent: any;
  maintanance: boolean;
  bussinessCount: number;
  contactInfo: ContactInfo;
  faqs: any[];
  status: "idle" | "loading" | "succeeded" | "failed";
  error: string | null;
  news: any[];
  loading: boolean;
}

const initialState: ConfigState = {
  websiteData: null,
  aboutContent: null,
  maintanance: false,
  bussinessCount: 0,
  contactInfo: {
    contactEmail: null,
    contactNumber: null,
  },
  faqs: [],
  status: "idle",
  error: null,
  news: [],
  loading: true,
};

const configSlice = createSlice({
  name: "config",
  initialState,
  reducers: {
    setWebsiteData: (state, action: PayloadAction<any>) => {
      state.websiteData = action.payload;
    },
    setAboutContent: (state, action: PayloadAction<any | null>) => {
      state.aboutContent = action.payload;
    },
    setMaintanance: (state, action: PayloadAction<boolean>) => {
      state.maintanance = action.payload;
    },
    setBussinessCount: (state, action: PayloadAction<number>) => {
      state.bussinessCount = action.payload;
    },
    setContactInfo: (state, action: PayloadAction<ContactInfo>) => {
      state.contactInfo = action.payload;
    },
    setFaqs: (state, action: PayloadAction<any[]>) => {
      state.faqs = action.payload;
    },
    setNews: (state, action: PayloadAction<any[]>) => {
      state.news = action.payload;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
  },
});

export const {
  setWebsiteData,
  setAboutContent,
  setMaintanance,
  setBussinessCount,
  setContactInfo,
  setFaqs,
  setNews,
  setLoading,
} = configSlice.actions;

export default configSlice.reducer;

export const fetchConfigData = () => async (dispatch: AppDispatch) => {
  try {
    const response = await api.get("/config");
    const {
      maintanance,
      aboutContent,
      count,
      contactEmail,
      contactNumber,
      faqs,
      news,
    } = response.data;
    dispatch(setAboutContent(aboutContent));
    dispatch(setMaintanance(maintanance));
    dispatch(setBussinessCount(count));
    dispatch(
      setContactInfo({
        contactEmail,
        contactNumber,
      })
    );
    dispatch(setFaqs(faqs));
    dispatch(setNews(news));
    dispatch(setLoading(false));
  } catch (error) {
    dispatch(setLoading(false));
    console.error("Error fetching config data:", error);
  }
};

export const updateAboutContent =
  (data: { aboutContent: string }) => async (dispatch: AppDispatch) => {
    try {
      const response = await api.put("/config/update", data);
      dispatch(setAboutContent(response.data.aboutContent));
    } catch (error) {
      console.error("Error updating about content:", error);
    }
  };

export const fetchFaqs = () => async (dispatch: AppDispatch) => {
  try {
    const response = await api.get("/config/faqs");
    dispatch(setFaqs(response.data));
  } catch (error) {
    console.error("Error fetching FAQs:", error);
  }
};
